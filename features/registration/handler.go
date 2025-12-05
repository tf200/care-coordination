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

func NewRegistrationHandler(rgstService RegistrationService, mdw *middleware.Middleware) *RegistrationHandler {
	return &RegistrationHandler{
		rgstService: rgstService,
		mdw:         mdw,
	}
}

func (h *RegistrationHandler) SetupRegistrationRoutes(router *gin.Engine) {
	registration := router.Group("/registrations")
	registration.Use(h.mdw.AuthMdw())
	registration.Use(h.mdw.PaginationMdw())

	registration.POST("", h.CreateRegistrationForm)
	registration.GET("", h.mdw.PaginationMdw(), h.ListRegistrationForms)
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
func (h *RegistrationHandler) CreateRegistrationForm(ctx *gin.Context) {
	var req CreateRegistrationFormRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.rgstService.CreateRegistrationForm(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// @Summary List registration forms
// @Description List all registration forms
// @Tags Registration
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search"
// @Param status query string false "Filter by status (pending, approved, rejected, in_review)"
// @Produce json
// @Success 200 {object} resp.PaginationResponse[ListRegistrationFormsResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /registrations [get]
func (h *RegistrationHandler) ListRegistrationForms(ctx *gin.Context) {
	var req ListRegistrationFormsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.rgstService.ListRegistrationForms(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, result)
}
