package locations

import (
	"care-cordination/lib/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	locationService LocationService
	mdw             *middleware.Middleware
}

func NewLocationHandler(
	locationService LocationService,
	mdw *middleware.Middleware,
) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
		mdw:             mdw,
	}
}

func (h *LocationHandler) SetupLocationRoutes(router *gin.Engine) {
	location := router.Group("/locations")

	location.POST("", h.mdw.AuthMdw(), h.CreateLocation)
	location.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListLocations)
	location.GET("/capacity-stats", h.mdw.AuthMdw(), h.GetLocationCapacityStats)
	location.PUT("/:id", h.mdw.AuthMdw(), h.UpdateLocation)
	location.DELETE("/:id", h.mdw.AuthMdw(), h.DeleteLocation)
}

// @Summary Create a location
// @Description Create a new location
// @Tags Location
// @Accept json
// @Produce json
// @Param location body CreateLocationRequest true "Location"
// @Success 200 {object} resp.SuccessResponse[CreateLocationResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /locations [post]
func (h *LocationHandler) CreateLocation(ctx *gin.Context) {
	var req CreateLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.locationService.CreateLocation(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Location created successfully"))
}

// @Summary List locations
// @Description List all locations with pagination and search
// @Tags Location
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by name, postal code, or address"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]ListLocationsResponse]]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /locations [get]
func (h *LocationHandler) ListLocations(ctx *gin.Context) {
	var req ListLocationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.locationService.ListLocations(ctx, &req)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Locations listed successfully"))
}

// @Summary Update a location
// @Description Update an existing location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param location body UpdateLocationRequest true "Location update data"
// @Success 200 {object} resp.SuccessResponse[UpdateLocationResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /locations/{id} [patch]
func (h *LocationHandler) UpdateLocation(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req UpdateLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.locationService.UpdateLocation(ctx, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Location updated successfully"))
}

// @Summary Delete a location (soft delete)
// @Description Soft delete a location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} resp.SuccessResponse[DeleteLocationResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /locations/{id} [delete]
func (h *LocationHandler) DeleteLocation(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.locationService.DeleteLocation(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Location deleted successfully"))
}

// @Summary Get location capacity statistics
// @Description Get total capacity, capacity used (clients in care), and free capacity across all locations
// @Tags Location
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetLocationCapacityStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /locations/capacity-stats [get]
func (h *LocationHandler) GetLocationCapacityStats(ctx *gin.Context) {
	result, err := h.locationService.GetLocationCapacityStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Location capacity statistics retrieved successfully"))
}
