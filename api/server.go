package api

// @title Care-Cordination API
// @version 1.0
// @description This is the Care-Cordination server API documentation.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email your-email@domain.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /

// @securityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @description Add 'Bearer ' prefix before your JWT token for authentication

// @Security Bearer
import (
	"care-cordination/docs"
	"care-cordination/features/attachments"
	"care-cordination/features/auth"
	"care-cordination/features/employee"
	"care-cordination/features/registration"
	"care-cordination/lib/logger"
	"care-cordination/lib/ratelimit"
	"context"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	httpServer          *http.Server
	router              *gin.Engine
	authHandler         *auth.AuthHandler
	employeeHandler     *employee.EmployeeHandler
	registrationHandler registration.RegistrationHandler
	attachmentsHandler  *attachments.AttachmentsHandler
	environment         string
	rateLimiter         ratelimit.RateLimiter
	logger              *logger.Logger
	addr                string
}

func NewServer(logger *logger.Logger,
	environment string, authHandler *auth.AuthHandler,
	employeeHandler *employee.EmployeeHandler,
	registrationHandler registration.RegistrationHandler,
	attachmentsHandler *attachments.AttachmentsHandler,
	rateLimiter ratelimit.RateLimiter, addr string) *Server {
	s := &Server{
		environment:         environment,
		authHandler:         authHandler,
		employeeHandler:     employeeHandler,
		registrationHandler: registrationHandler,
		attachmentsHandler:  attachmentsHandler,
		rateLimiter:         rateLimiter,
		logger:              logger,
		addr:                addr,
	}
	s.setupRoutes(logger)
	return s
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) setupRoutes(logger *logger.Logger) {
	s.setupSwagger()
	gin.SetMode(func() string {
		if s.environment == "production" {
			return gin.ReleaseMode
		}
		return gin.DebugMode
	}())
	router := gin.New()
	router.Use(ginzap.GinzapWithConfig(logger.Logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: "2006-01-02 15:04:05",
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			if requestID, ok := c.Get("X-Request-Id"); ok {
				fields = append(fields, zap.String("request_id", requestID.(string)))
			}
			return fields
		}),
	}))
	router.Use(ginzap.RecoveryWithZap(logger.Logger, true))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.authHandler.SetupAuthRoutes(router, s.rateLimiter, logger)
	s.employeeHandler.SetupEmployeeRoutes(router)
	s.registrationHandler.SetupRegistrationRoutes(router)
	s.attachmentsHandler.SetupAttachmentsRoutes(router)
	s.router = router
}

func (s *Server) setupSwagger() {
	docs.SwaggerInfo.Title = "Care-Cordination API"
	docs.SwaggerInfo.Description = "This is the Care-Cordination server API documentation."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = s.addr

	// Set scheme based on environment
	if s.environment == "production" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
}
