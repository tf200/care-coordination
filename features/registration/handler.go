package registration

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	rgstService RegistrationService
	mdw         *middleware.Middleware
}

func NewRegistrationHandler(rgstService RegistrationService, mdw *middleware.Middleware) RegistrationHandler {
	return RegistrationHandler{
		rgstService: rgstService,
		mdw:         mdw,
	}
}

func (h *RegistrationHandler) SetupRegistrationRoutes(router *gin.Engine) {
	registration := router.Group("/registrations")

	registration.POST("", h.mdw.AuthMiddleware(), h.CreateRegistrationForm)
	registration.GET("", h.mdw.AuthMiddleware(), h.ListRegistrationForms)
}

func (h *RegistrationHandler) CreateRegistrationForm(c *gin.Context) {
	var req CreateRegistrationFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp.ErrorResponse(ErrInvalidRequest))
		return
	}

	result, err := h.rgstService.CreateRegistrationForm(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.ErrorResponse(ErrInternal))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *RegistrationHandler) ListRegistrationForms(c *gin.Context) {
	result, err := h.rgstService.ListRegistrationForms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.ErrorResponse(ErrInternal))
		return
	}
	c.JSON(http.StatusOK, result)
}
