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

// @Summary Create a registration form
// @Description Create a new registration form
// @Tags Registration
// @Accept json
// @Produce json
// @Param registration body CreateRegistrationFormRequest true "Registration Form"
// @Success 200 {object} CreateRegistrationFormResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /registrations [post]
func (h *RegistrationHandler) CreateRegistrationForm(c *gin.Context) {
	var req CreateRegistrationFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.rgstService.CreateRegistrationForm(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary List registration forms
// @Description List all registration forms
// @Tags Registration
// @Accept json
// @Produce json
// @Success 200 {object} []ListRegistrationFormsResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /registrations [get]
func (h *RegistrationHandler) ListRegistrationForms(c *gin.Context) {
	result, err := h.rgstService.ListRegistrationForms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	c.JSON(http.StatusOK, result)
}
