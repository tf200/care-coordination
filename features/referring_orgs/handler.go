package referringOrgs

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReferringOrgHandler struct {
	service ReferringOrgService
	mdw     *middleware.Middleware
}

func NewReferringOrgHandler(
	service ReferringOrgService,
	mdw *middleware.Middleware,
) *ReferringOrgHandler {
	return &ReferringOrgHandler{
		service: service,
		mdw:     mdw,
	}
}

func (h *ReferringOrgHandler) SetupReferringOrgRoutes(router *gin.Engine) {
	orgs := router.Group("/referring-orgs")

	orgs.POST("", h.mdw.AuthMdw(), h.CreateReferringOrg)
	orgs.GET("/stats", h.mdw.AuthMdw(), h.GetReferringOrgStats)
	orgs.GET("", h.mdw.AuthMdw(), h.ListReferringOrgs)
	orgs.PUT("/:id", h.mdw.AuthMdw(), h.UpdateReferringOrg)
}

// @Summary Create a new referring organization
// @Description Create a new referring organization with the provided details
// @Tags referring-orgs
// @Accept json
// @Produce json
// @Param request body CreateReferringOrgRequest true "Referring Organization data"
// @Success 201 {object} CreateReferringOrgResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /referring-orgs [post]
func (h *ReferringOrgHandler) CreateReferringOrg(ctx *gin.Context) {
	var req CreateReferringOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.service.CreateReferringOrg(ctx.Request.Context(), &req)
	if err != nil {
		if err == ErrInternal {
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
			return
		}
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusCreated, resp.Success(result, "Referring organization created successfully"))
}

// @Summary List referring organizations
// @Description Get a paginated list of referring organizations with optional search
// @Tags referring-orgs
// @Accept json
// @Produce json
// @Param search query string false "Search term for name, contact person, or email"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Items per page (default: 10)"
// @Success 200 {object} resp.PaginationResponse[ListReferringOrgsResponse]
// @Failure 500 {object} resp.ErrorResponse
// @Router /referring-orgs [get]
func (h *ReferringOrgHandler) ListReferringOrgs(ctx *gin.Context) {
	var req ListReferringOrgsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.service.ListReferringOrgs(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Referring organizations listed successfully"))
}

// @Summary Update a referring organization
// @Description Update an existing referring organization with partial data
// @Tags referring-orgs
// @Accept json
// @Produce json
// @Param id path string true "Referring Organization ID"
// @Param request body UpdateReferringOrgRequest true "Referring Organization update data"
// @Success 200 {object} UpdateReferringOrgResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /referring-orgs/{id} [put]
func (h *ReferringOrgHandler) UpdateReferringOrg(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req UpdateReferringOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.service.UpdateReferringOrg(ctx.Request.Context(), id, &req)
	if err != nil {
		if err == ErrInternal {
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
			return
		}
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Referring organization updated successfully"))
}

// @Summary Get referring organization statistics
// @Description Get comprehensive statistics for referring organizations including total orgs, orgs with in-care/waitlist clients, and total clients referred
// @Tags referring-orgs
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetReferringOrgStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /referring-orgs/stats [get]
func (h *ReferringOrgHandler) GetReferringOrgStats(ctx *gin.Context) {
	result, err := h.service.GetReferringOrgStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Referring organization statistics retrieved successfully"))
}
