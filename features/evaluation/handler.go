package evaluation

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EvaluationHandler struct {
	service EvaluationService
	mdw     *middleware.Middleware
}

func NewEvaluationHandler(service EvaluationService, mdw *middleware.Middleware) *EvaluationHandler {
	return &EvaluationHandler{
		service: service,
		mdw:     mdw,
	}
}

func (h *EvaluationHandler) SetupEvaluationRoutes(router *gin.Engine) {
	ev := router.Group("/evaluations")
	ev.Use(h.mdw.AuthMdw())
	ev.Use(h.mdw.PaginationMdw())

	ev.POST("", h.CreateEvaluation)
	ev.GET("/critical", h.GetCritical)
	ev.GET("/scheduled", h.GetScheduled)
	ev.GET("/recent", h.GetRecent)
	ev.GET("/history/:clientId", h.GetEvaluationHistory)
}

// @Summary Create a client evaluation
// @Description Record progress logs for all current client goals and schedule the next evaluation.
// @Tags Evaluation
// @Accept json
// @Produce json
// @Param request body CreateEvaluationRequest true "Evaluation Details"
// @Success 200 {object} resp.SuccessResponse[CreateEvaluationResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations [post]
func (h *EvaluationHandler) CreateEvaluation(c *gin.Context) {
	var req CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.service.CreateEvaluation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Evaluation created successfully"))
}

// @Summary Get evaluation history for a client
// @Description Retrieve all past evaluations and goal progress logs for a specific client.
// @Tags Evaluation
// @Produce json
// @Param clientId path string true "Client ID"
// @Success 200 {object} resp.SuccessResponse[[]EvaluationHistoryItem]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/history/{clientId} [get]
func (h *EvaluationHandler) GetEvaluationHistory(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, resp.Error(nil)) // nil error is not great, but following format
		return
	}

	result, err := h.service.GetEvaluationHistory(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "History retrieved successfully"))
}

// @Summary Get critical evaluations (Dashboard)
// @Description List evaluations due within the next 7 days or overdue.
// @Tags Evaluation
// @Produce json
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[UpcomingEvaluationDTO]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/critical [get]
func (h *EvaluationHandler) GetCritical(c *gin.Context) {
	result, err := h.service.GetCriticalEvaluations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Critical evaluations retrieved successfully"))
}

// @Summary Get scheduled evaluations (Dashboard)
// @Description List evaluations scheduled between 8 and 30 days from now.
// @Tags Evaluation
// @Produce json
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[UpcomingEvaluationDTO]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/scheduled [get]
func (h *EvaluationHandler) GetScheduled(c *gin.Context) {
	result, err := h.service.GetScheduledEvaluations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Scheduled evaluations retrieved successfully"))
}

// @Summary Get recent evaluations (Dashboard)
// @Description List the last 20 evaluations submitted across all clients.
// @Tags Evaluation
// @Produce json
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[GlobalRecentEvaluationDTO]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/recent [get]
func (h *EvaluationHandler) GetRecent(c *gin.Context) {
	result, err := h.service.GetRecentEvaluations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Recent evaluations retrieved successfully"))
}
