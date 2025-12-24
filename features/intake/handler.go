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
	intake.GET("/stats", h.GetIntakeStats)
	intake.GET("/:id", h.GetIntakeForm)
	intake.PUT("/:id", h.UpdateIntakeForm)
}

// @Summary Create an intake form
// @Description Create a new intake form
// @Tags Intake
// @Accept json
// @Produce json
// @Param intake body CreateIntakeFormRequest true "Intake Form"
// @Success 200 {object} resp.SuccessResponse[CreateIntakeFormResponse]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Intake form created successfully"))
}

// @Summary List intake forms
// @Description List all intake forms
// @Tags Intake
// @Accept json
// @Produce json
// @Param search query string false "Search by client first name or last name"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[ListIntakeFormsResponse]]
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
	ctx.JSON(http.StatusOK, resp.Success(result, "Intake forms listed successfully"))
}

// @Summary Get an intake form
// @Description Get an intake form by ID with details
// @Tags Intake
// @Produce json
// @Param id path string true "Intake Form ID"
// @Success 200 {object} resp.SuccessResponse[GetIntakeFormResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /intakes/{id} [get]
func (h *IntakeHandler) GetIntakeForm(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.intakeService.GetIntakeForm(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Intake form retrieved successfully"))
}

// @Summary Update an intake form
// @Description Update an existing intake form
// @Tags Intake
// @Accept json
// @Produce json
// @Param id path string true "Intake Form ID"
// @Param intake body UpdateIntakeFormRequest true "Intake Form Update"
// @Success 200 {object} resp.SuccessResponse[UpdateIntakeFormResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /intakes/{id} [put]
func (h *IntakeHandler) UpdateIntakeForm(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req UpdateIntakeFormRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.intakeService.UpdateIntakeForm(ctx, id, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Intake form updated successfully"))
}

// @Summary Get intake statistics
// @Description Get total count, pending count, and conversion percentage of intake forms
// @Tags Intake
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetIntakeStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /intakes/stats [get]
func (h *IntakeHandler) GetIntakeStats(ctx *gin.Context) {
	result, err := h.intakeService.GetIntakeStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Intake statistics retrieved successfully"))
}
