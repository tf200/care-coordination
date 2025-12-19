package middleware

import (
	"care-cordination/lib/ratelimit"
	"care-cordination/lib/resp"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

// RateLimitMiddleware creates a middleware for rate limiting login requests
func (m *Middleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip rate limiting if limiter is nil
		if m.rateLimiter == nil {
			ctx.Next()
			return
		}

		// Get client IP (handles X-Forwarded-For for proxies)
		ip := getClientIP(ctx)

		// Check IP-based rate limit
		result, err := m.rateLimiter.CheckIPLimit(ctx, ip)
		if err != nil {
			// Log error but don't block the request (fail open)
			m.logger.Error(ctx, "RateLimitMiddleware", "IP rate limit check failed",
				zap.Error(err), zap.String("ip", ip))
			ctx.Next()
			return
		}

		// Set rate limit headers
		ctx.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		ctx.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			// Log rate limit violation
			m.logger.Warn(ctx, "RateLimitMiddleware", "Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", ctx.Request.URL.Path),
				zap.Duration("retry_after", result.RetryAfter))

			// Set Retry-After header
			ctx.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))

			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, resp.Error(ErrRateLimitExceeded))
			return
		}

		ctx.Next()
	}
}

// LoginRateLimitMiddleware creates a middleware specifically for login endpoint
// It checks both IP and email-based rate limits
func (m *Middleware) LoginRateLimitMiddleware(
	limiter ratelimit.RateLimiter,
	logger *zap.Logger,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip rate limiting if limiter is nil
		if limiter == nil {
			ctx.Next()
			return
		}

		// Get client IP
		ip := getClientIP(ctx)

		// Check IP-based rate limit first
		ipResult, err := limiter.CheckIPLimit(ctx, ip)
		if err != nil {
			logger.Error("IP rate limit check failed",
				zap.Error(err),
				zap.String("ip", ip))
			// Fail open - continue to next check
		} else {
			// Set IP-based rate limit headers
			ctx.Header("X-RateLimit-IP-Limit", strconv.Itoa(ipResult.Limit))
			ctx.Header("X-RateLimit-IP-Remaining", strconv.Itoa(ipResult.Remaining))
			ctx.Header("X-RateLimit-IP-Reset", strconv.FormatInt(ipResult.ResetAt.Unix(), 10))

			if !ipResult.Allowed {
				logger.Warn("IP rate limit exceeded",
					zap.String("ip", ip),
					zap.String("user_agent", ctx.Request.UserAgent()),
					zap.Duration("retry_after", ipResult.RetryAfter))

				ctx.Header("Retry-After", strconv.Itoa(int(ipResult.RetryAfter.Seconds())))
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, resp.Error(ErrRateLimitIP))
				return
			}
		}

		// Parse request body to get email for email-based rate limiting
		// We need to peek at the body without consuming it
		var loginReq struct {
			Email string `json:"email"`
		}

		// Bind JSON but allow continuing even if it fails
		// (the actual handler will do proper validation)
		// Use ShouldBindBodyWith to cache the body so it can be re-read by the handler
		if err := ctx.ShouldBindBodyWith(&loginReq, binding.JSON); err == nil &&
			loginReq.Email != "" {
			// Check email-based rate limit
			emailResult, err := limiter.CheckEmailLimit(ctx, loginReq.Email)
			if err != nil {
				logger.Error("Email rate limit check failed",
					zap.Error(err),
					zap.String("email", loginReq.Email))
				// Fail open - continue to handler
			} else {
				// Set email-based rate limit headers
				ctx.Header("X-RateLimit-Email-Limit", strconv.Itoa(emailResult.Limit))
				ctx.Header("X-RateLimit-Email-Remaining", strconv.Itoa(emailResult.Remaining))
				ctx.Header("X-RateLimit-Email-Reset", strconv.FormatInt(emailResult.ResetAt.Unix(), 10))

				if !emailResult.Allowed {
					logger.Warn("Email rate limit exceeded",
						zap.String("email", loginReq.Email),
						zap.String("ip", ip),
						zap.Duration("retry_after", emailResult.RetryAfter))

					ctx.Header("Retry-After", strconv.Itoa(int(emailResult.RetryAfter.Seconds())))
					ctx.AbortWithStatusJSON(http.StatusTooManyRequests, resp.Error(ErrRateLimitEmail))
					return
				}
			}
		}

		ctx.Next()
	}
}

// getClientIP extracts the real client IP from the request
// Handles X-Forwarded-For header for requests behind proxies/load balancers
func getClientIP(ctx *gin.Context) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := ctx.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2, ...)
		// Take the first one (the original client)
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header (alternative to X-Forwarded-For)
	if xri := ctx.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return ctx.ClientIP()
}

// SetRateLimitContext stores rate limit information in the context
// This can be used by handlers to reset limits on successful login
func SetRateLimitContext(ctx *gin.Context, email string) {
	ctx.Set("rate_limit_email", email)
}

// GetRateLimitEmail retrieves the email from rate limit context
func GetRateLimitEmail(ctx *gin.Context) (string, bool) {
	email, exists := ctx.Get("rate_limit_email")
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}
