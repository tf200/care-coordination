package middleware

import (
	"bytes"
	"care-cordination/lib/audit"
	"care-cordination/lib/util"
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuditAction represents an audit action type
type AuditAction string

const (
	ActionRead   AuditAction = "read"
	ActionCreate AuditAction = "create"
	ActionUpdate AuditAction = "update"
	ActionDelete AuditAction = "delete"
)

// AuditStatus represents the status of an audit entry
type AuditStatus string

const (
	StatusSuccess AuditStatus = "success"
	StatusFailure AuditStatus = "failure"
)

// AuditEntry represents a single audit log entry for the middleware
type AuditEntry struct {
	UserID        string
	EmployeeID    string
	ClientID      string // NEN7510: Track which client's data was accessed
	Action        AuditAction
	ResourceType  string
	ResourceID    string
	OldValue      any
	NewValue      any
	IPAddress     string
	UserAgent     string
	RequestID     string
	Status        AuditStatus
	FailureReason string
}

// AuditLogger is an interface for logging audit entries
// This avoids import cycle between middleware and audit packages
type AuditLogger interface {
	LogEntry(ctx context.Context, entry AuditEntry) error
}

// responseWriter wraps gin.ResponseWriter to capture status code
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w gin.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     200,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

var routeResourceMap = map[string]string{
	"/attachments":        audit.ResourceTypeAttachment,
	"/audit":              audit.ResourceTypeAudit,
	"/calendar":           audit.ResourceTypeCalendar,
	"/clients":            audit.ResourceTypeClient,
	"/employees":          audit.ResourceTypeEmployee,
	"/evaluations":        audit.ResourceTypeEvaluation,
	"/incidents":          audit.ResourceTypeIncident,
	"/intake-forms":       audit.ResourceTypeIntakeForm,
	"/locations":          audit.ResourceTypeLocation,
	"/location-transfers": audit.ResourceTypeLocationTransfer,
	"/notifications":      audit.ResourceTypeNotification,
	"/rbac":               audit.ResourceTypeRBAC,
	"/referring-orgs":     audit.ResourceTypeReferringOrg,
	"/registrations":      audit.ResourceTypeRegistration,
}

// AuditMdw returns a middleware that logs all requests to the audit service
func (m *Middleware) AuditMdw() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip audit logging for certain paths
		path := ctx.Request.URL.Path
		if shouldSkipAudit(path) {
			ctx.Next()
			return
		}

		// Wrap response writer to capture status
		rw := newResponseWriter(ctx.Writer)
		ctx.Writer = rw

		// Process request
		ctx.Next()

		// Log after request completes (fire and forget)
		go m.logAuditEntry(ctx.Copy(), rw.statusCode, path, ctx.Request.Method)
	}
}

// shouldSkipAudit returns true for paths that shouldn't be audited
func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/swagger",
		"/health",
		"/metrics",
		"/auth/login",   // Login is handled separately in auth service
		"/auth/logout",  // Logout is handled separately in auth service
		"/auth/refresh", // Token refresh doesn't need audit
		"/audit",        // Don't audit audit log access (avoid recursion)
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// httpMethodToAction maps HTTP methods to audit actions
func httpMethodToAction(method string) AuditAction {
	switch method {
	case "GET":
		return ActionRead
	case "POST":
		return ActionCreate
	case "PUT", "PATCH":
		return ActionUpdate
	case "DELETE":
		return ActionDelete
	default:
		return ActionRead
	}
}

// extractResourceInfo extracts resource type and ID from URL path
func extractResourceInfo(path string) (resourceType, resourceID string) {
	// Check for exact match or prefix match
	for route, rt := range routeResourceMap {
		if strings.HasPrefix(path, route) {
			resourceType = rt
			// Extract ID if present after the route
			remaining := strings.TrimPrefix(path, route)
			if len(remaining) > 0 && remaining[0] == '/' {
				resourceID = strings.TrimPrefix(remaining, "/")
			}
			return
		}
	}
	return
}

// logAuditEntry creates an audit log entry for the request
func (m *Middleware) logAuditEntry(ctx *gin.Context, statusCode int, path, method string) {
	if m.auditLogger == nil {
		return
	}

	action := httpMethodToAction(method)
	resourceType, resourceID := extractResourceInfo(path)

	// Determine status
	status := StatusSuccess
	var failureReason string
	if statusCode >= 400 {
		status = StatusFailure
		if statusCode == 401 {
			failureReason = "unauthorized"
		} else if statusCode == 403 {
			failureReason = "forbidden"
		} else if statusCode == 404 {
			failureReason = "not_found"
		} else if statusCode >= 500 {
			failureReason = "server_error"
		} else {
			failureReason = "client_error"
		}
	}

	entry := AuditEntry{
		UserID:        util.GetUserID(ctx),
		EmployeeID:    util.GetEmployeeID(ctx),
		ClientID:      util.GetClientID(ctx),
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		IPAddress:     util.GetIPAddress(ctx),
		UserAgent:     util.GetUserAgent(ctx),
		RequestID:     util.GetRequestID(ctx),
		Status:        status,
		FailureReason: failureReason,
	}

	// Log silently - don't disrupt the request flow
	_ = m.auditLogger.LogEntry(ctx, entry)
}
