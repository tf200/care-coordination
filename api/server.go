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
	"care-cordination/features/audit"
	"care-cordination/features/auth"
	"care-cordination/features/calendar"
	"care-cordination/features/client"
	"care-cordination/features/employee"
	"care-cordination/features/evaluation"
	"care-cordination/features/incident"
	"care-cordination/features/intake"
	locTransfer "care-cordination/features/location_transfer"
	"care-cordination/features/locations"
	"care-cordination/features/middleware"
	"care-cordination/features/notification"
	"care-cordination/features/rbac"
	referringOrgs "care-cordination/features/referring_orgs"
	"care-cordination/features/registration"
	"care-cordination/lib/logger"
	"care-cordination/lib/ratelimit"
	"care-cordination/lib/websocket"
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	// handlers
	authHandler         *auth.AuthHandler
	locationHandler     *locations.LocationHandler
	employeeHandler     *employee.EmployeeHandler
	registrationHandler *registration.RegistrationHandler
	intakeHandler       *intake.IntakeHandler
	incidentHandler     *incident.IncidentHandler
	attachmentsHandler  *attachments.AttachmentsHandler
	clientHandler       *client.ClientHandler
	referringOrgHandler *referringOrgs.ReferringOrgHandler
	locTransferHandler  *locTransfer.LocTransferHandler
	evaluationHandler   *evaluation.EvaluationHandler
	rbacHandler         *rbac.RBACHandler
	calendarHandler     *calendar.CalendarHandler
	notificationHandler *notification.NotificationHandler
	auditHandler        *audit.AuditHandler
	wsHub               *websocket.Hub

	environment string
	rateLimiter ratelimit.RateLimiter
	logger      logger.Logger
	addr        string
	url         string
}

func NewServer(
	logger logger.Logger,
	environment string,
	authHandler *auth.AuthHandler,
	employeeHandler *employee.EmployeeHandler,
	registrationHandler *registration.RegistrationHandler,
	attachmentsHandler *attachments.AttachmentsHandler,
	locationHandler *locations.LocationHandler,
	intakeHandler *intake.IntakeHandler,
	incidentHandler *incident.IncidentHandler,
	clientHandler *client.ClientHandler,
	referringOrgHandler *referringOrgs.ReferringOrgHandler,
	locTransferHandler *locTransfer.LocTransferHandler,
	rbacHandler *rbac.RBACHandler,
	evaluationHandler *evaluation.EvaluationHandler,
	calendarHandler *calendar.CalendarHandler,
	notificationHandler *notification.NotificationHandler,
	auditHandler *audit.AuditHandler,
	wsHub *websocket.Hub,
	rateLimiter ratelimit.RateLimiter, addr string, url string) *Server {
	s := &Server{
		environment:         environment,
		authHandler:         authHandler,
		employeeHandler:     employeeHandler,
		registrationHandler: registrationHandler,
		attachmentsHandler:  attachmentsHandler,
		rateLimiter:         rateLimiter,
		locationHandler:     locationHandler,
		intakeHandler:       intakeHandler,
		incidentHandler:     incidentHandler,
		clientHandler:       clientHandler,
		referringOrgHandler: referringOrgHandler,
		locTransferHandler:  locTransferHandler,
		rbacHandler:         rbacHandler,
		evaluationHandler:   evaluationHandler,
		calendarHandler:     calendarHandler,
		notificationHandler: notificationHandler,
		auditHandler:        auditHandler,
		wsHub:               wsHub,
		logger:              logger,
		addr:                addr,
		url:                 url,
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

func (s *Server) setupRoutes(logger logger.Logger) {
	s.setupSwagger()
	gin.SetMode(func() string {
		if s.environment == "production" {
			return gin.ReleaseMode
		}
		return gin.DebugMode
	}())
	router := gin.New()

	// CORS middleware - must be before other middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, specify exact origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Request ID middleware - must be before ginzap for logging
	router.Use(middleware.RequestIDMiddleware())

	router.Use(ginzap.GinzapWithConfig(logger.ZapLogger(), &ginzap.Config{
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
	router.Use(ginzap.RecoveryWithZap(logger.ZapLogger(), true))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.authHandler.SetupAuthRoutes(router, s.rateLimiter)
	s.employeeHandler.SetupEmployeeRoutes(router)
	s.registrationHandler.SetupRegistrationRoutes(router)
	s.attachmentsHandler.SetupAttachmentsRoutes(router)
	s.locationHandler.SetupLocationRoutes(router)
	s.intakeHandler.SetupIntakeRoutes(router)
	s.incidentHandler.SetupIncidentRoutes(router)
	s.clientHandler.SetupClientRoutes(router)
	s.referringOrgHandler.SetupReferringOrgRoutes(router)
	s.locTransferHandler.SetupLocTransferRoutes(router)
	s.rbacHandler.SetupRBACRoutes(router)
	s.evaluationHandler.SetupEvaluationRoutes(router)
	s.calendarHandler.SetupRoutes(router)
	s.notificationHandler.SetupRoutes(router)
	s.auditHandler.SetupAuditRoutes(router)
	s.router = router
}

func (s *Server) setupSwagger() {
	docs.SwaggerInfo.Title = "Care-Cordination API"
	docs.SwaggerInfo.Description = "This is the Care-Cordination server API documentation."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = s.url

	// Set scheme based on environment
	if s.environment == "production" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
}
