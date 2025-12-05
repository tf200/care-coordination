package intake

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IntakeHandler struct {
	intakeService IntakeService
	mdw           *middleware.Middleware
}

func NewIntakeHandler(intakeService IntakeService, mdw *middleware.Middleware) *IntakeHandler {
	return &IntakeHandler{
		intakeService: intakeService,
		mdw:           mdw,
	}
}

func (h *IntakeHandler) SetupIntakeRoutes(router *gin.Engine) {
	intake := router.Group("/intakes")
	intake.Use(h.mdw.AuthMdw())
	intake.Use(h.mdw.PaginationMdw())

	intake.POST("", h.CreateIntakeForm)
	intake.GET("", h.ListIntakeForms)
}

// @Summary Create an intake form
// @Description Create a new intake form
// @Tags Intake
// @Accept json
// @Produce json
// @Param intake body CreateIntakeFormRequest true "Intake Form"
// @Success 200 {object} CreateIntakeFormResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /intakes [post]
func (h *IntakeHandler) CreateIntakeForm(ctx *gin.Context) {
	var req CreateIntakeFormRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.intakeService.CreateIntakeForm(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// @Summary List intake forms
// @Description List all intake forms
// @Tags Intake
// @Accept json
// @Produce json
// @Param search query string false "Search by client first name or last name"
// @Success 200 {object} resp.PaginationResponse[ListIntakeFormsResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /intakes [get]
func (h *IntakeHandler) ListIntakeForms(ctx *gin.Context) {
	var req ListIntakeFormsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}
	if req.Search == nil {
		empty := ""
		req.Search = &empty
	}
	result, err := h.intakeService.ListIntakeForms(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, result)
}
