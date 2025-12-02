package referring_orgs

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/logger"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReferringOrgHandler struct {
	service ReferringOrgService
	mdw     *middleware.Middleware
}

func NewReferringOrgHandler(service ReferringOrgService, mdw *middleware.Middleware) *ReferringOrgHandler {
	return &ReferringOrgHandler{
		service: service,
		mdw:     mdw,
	}
}

func (h *ReferringOrgHandler) SetupReferringOrgRoutes(router *gin.Engine, logger *logger.Logger) {
	orgs := router.Group("/api/referring-orgs")

	orgs.POST("", h.mdw.AuthMdw(), h.CreateReferringOrg)
	orgs.GET("", h.mdw.AuthMdw(), h.ListReferringOrgs)
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
// @Router /api/referring-orgs [post]
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

	ctx.JSON(http.StatusCreated, result)
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
// @Router /api/referring-orgs [get]
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

	ctx.JSON(http.StatusOK, result)
}
