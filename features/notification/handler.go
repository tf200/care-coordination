package notification

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"care-cordination/lib/token"
	"care-cordination/lib/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		return true
	},
}

// NotificationHandler handles notification HTTP and WebSocket endpoints
type NotificationHandler struct {
	service       NotificationService
	hub           *websocket.Hub
	ticketManager *websocket.TicketManager
	tokenManager  token.TokenManager
	mdw           *middleware.Middleware
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(
	service NotificationService,
	hub *websocket.Hub,
	ticketManager *websocket.TicketManager,
	tokenManager token.TokenManager,
	mdw *middleware.Middleware,
) *NotificationHandler {
	return &NotificationHandler{
		service:       service,
		hub:           hub,
		ticketManager: ticketManager,
		tokenManager:  tokenManager,
		mdw:           mdw,
	}
}

// SetupRoutes registers notification routes
func (h *NotificationHandler) SetupRoutes(router *gin.Engine) {
	// REST API routes (require auth)
	notifications := router.Group("/notifications")
	notifications.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListNotifications)
	notifications.GET("/unread-count", h.mdw.AuthMdw(), h.GetUnreadCount)
	notifications.PATCH("/:id/read", h.mdw.AuthMdw(), h.MarkAsRead)
	notifications.PATCH("/read-all", h.mdw.AuthMdw(), h.MarkAllAsRead)
	notifications.DELETE("/:id", h.mdw.AuthMdw(), h.DeleteNotification)

	// WebSocket auth ticket endpoint
	router.POST("/ws/auth", h.mdw.AuthMdw(), h.CreateWSTicket)

	// WebSocket connection endpoint (uses ticket auth)
	router.GET("/ws/notifications", h.HandleWebSocket)
}

// @Summary List notifications
// @Description List notifications for the current user with optional filtering
// @Tags Notifications
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param is_read query bool false "Filter by read status"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]NotificationResponse]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Security Bearer
// @Router /notifications [get]
func (h *NotificationHandler) ListNotifications(ctx *gin.Context) {
	var req ListNotificationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.service.List(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Notifications listed successfully"))
}

// @Summary Get unread notification count
// @Description Get the count of unread notifications for the current user
// @Tags Notifications
// @Produce json
// @Success 200 {object} resp.SuccessResponse[UnreadCountResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Security Bearer
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(ctx *gin.Context) {
	count, err := h.service.GetUnreadCount(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(UnreadCountResponse{Count: count}, "Unread count retrieved"))
}

// @Summary Mark notification as read
// @Description Mark a single notification as read
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Security Bearer
// @Router /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsRead(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	err := h.service.MarkAsRead(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Notification marked as read"))
}

// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the current user
// @Tags Notifications
// @Produce json
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Security Bearer
// @Router /notifications/read-all [patch]
func (h *NotificationHandler) MarkAllAsRead(ctx *gin.Context) {
	err := h.service.MarkAllAsRead(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("All notifications marked as read"))
}

// @Summary Delete notification
// @Description Delete a notification
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Security Bearer
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	err := h.service.Delete(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Notification deleted"))
}

// @Summary Create WebSocket auth ticket
// @Description Exchange JWT for a one-time WebSocket connection ticket
// @Tags Notifications
// @Produce json
// @Success 200 {object} resp.SuccessResponse[WSAuthResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Security Bearer
// @Router /ws/auth [post]
func (h *NotificationHandler) CreateWSTicket(ctx *gin.Context) {
	// User is already authenticated via middleware, get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrUnauthorized))
		return
	}

	// Create one-time ticket
	ticket, err := h.ticketManager.CreateTicket(ctx, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(WSAuthResponse{Ticket: ticket}, "Ticket created"))
}

// HandleWebSocket handles WebSocket connection upgrade
// @Summary Connect to WebSocket
// @Description Establish WebSocket connection for real-time notifications
// @Tags Notifications
// @Param ticket query string true "One-time auth ticket from /ws/auth"
// @Success 101 "Switching Protocols"
// @Failure 401 {object} resp.ErrorResponse
// @Router /ws/notifications [get]
func (h *NotificationHandler) HandleWebSocket(ctx *gin.Context) {
	// Get ticket from query param
	ticket := ctx.Query("ticket")
	if ticket == "" {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrMissingToken))
		return
	}

	// Validate and consume ticket (one-time use)
	userID, err := h.ticketManager.ValidateTicket(ctx, ticket)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInvalidTicket))
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		// Upgrade handles the response, just return
		return
	}

	// Create client and register with hub
	client := websocket.NewClient(h.hub, conn, userID)

	// Set message handler for client-to-server messages
	client.SetMessageHandler(func(c *websocket.Client, msg *websocket.ClientMessage) {
		h.handleClientMessage(ctx, c, msg)
	})

	h.hub.Register(client)

	// Send connected message
	client.SendMessage(&websocket.Message{
		Type: websocket.MessageTypeConnected,
	})

	// Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}

// handleClientMessage handles messages from WebSocket clients
func (h *NotificationHandler) handleClientMessage(ctx *gin.Context, client *websocket.Client, msg *websocket.ClientMessage) {
	switch msg.Type {
	case websocket.MessageTypeMarkRead:
		if msg.Payload != "" {
			// Mark specific notification as read
			_ = h.service.MarkAsRead(ctx, msg.Payload)
		}
	case websocket.MessageTypeMarkAllRead:
		// Mark all as read
		_ = h.service.MarkAllAsRead(ctx)
	case websocket.MessageTypePong:
		// Client responded to ping, connection is alive
	}
}
