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
	ev.GET("/last/:clientId", h.GetLastEvaluation)

	// Draft endpoints
	ev.POST("/drafts", h.SaveDraft)
	ev.GET("/drafts", h.GetDrafts)
	ev.GET("/drafts/:id", h.GetDraftById)
	ev.POST("/drafts/:id/submit", h.SubmitDraft)
	ev.DELETE("/drafts/:id", h.DeleteDraft)
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

// @Summary Get last evaluation for a client
// @Description Retrieve the most recent evaluation with all goal progress logs for a specific client.
// @Tags Evaluation
// @Produce json
// @Param clientId path string true "Client ID"
// @Success 200 {object} resp.SuccessResponse[LastEvaluationDTO]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/last/{clientId} [get]
func (h *EvaluationHandler) GetLastEvaluation(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, resp.Error(nil))
		return
	}

	result, err := h.service.GetLastEvaluation(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, resp.Error(nil))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Last evaluation retrieved successfully"))
}

// @Summary Save or update a draft evaluation
// @Description Create a new draft or update an existing draft evaluation for a client.
// @Tags Evaluation
// @Accept json
// @Produce json
// @Param request body SaveDraftRequest true "Draft Evaluation Details"
// @Success 200 {object} resp.SuccessResponse[SaveDraftResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/drafts [post]
func (h *EvaluationHandler) SaveDraft(c *gin.Context) {
	var req SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.service.SaveDraft(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Draft saved successfully"))
}

// @Summary Get all draft evaluations
// @Description Retrieve all draft evaluations for the logged-in coordinator.
// @Tags Evaluation
// @Produce json
// @Param coordinatorId query string true "Coordinator ID"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[DraftEvaluationListItemDTO]]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/drafts [get]
func (h *EvaluationHandler) GetDrafts(c *gin.Context) {

	result, err := h.service.GetDrafts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Drafts retrieved successfully"))
}

// @Summary Get a specific draft evaluation
// @Description Retrieve a draft evaluation by ID with all progress logs.
// @Tags Evaluation
// @Produce json
// @Param id path string true "Draft Evaluation ID"
// @Success 200 {object} resp.SuccessResponse[DraftEvaluationDTO]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/drafts/{id} [get]
func (h *EvaluationHandler) GetDraftById(c *gin.Context) {
	evaluationID := c.Param("id")
	if evaluationID == "" {
		c.JSON(http.StatusBadRequest, resp.Error(nil))
		return
	}

	result, err := h.service.GetDraft(c.Request.Context(), evaluationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, resp.Error(nil))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Draft retrieved successfully"))
}

// @Summary Submit a draft evaluation
// @Description Convert a draft evaluation to submitted status and update client's next evaluation date.
// @Tags Evaluation
// @Produce json
// @Param id path string true "Draft Evaluation ID"
// @Success 200 {object} resp.SuccessResponse[CreateEvaluationResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/drafts/{id}/submit [post]
func (h *EvaluationHandler) SubmitDraft(c *gin.Context) {
	evaluationID := c.Param("id")
	if evaluationID == "" {
		c.JSON(http.StatusBadRequest, resp.Error(nil))
		return
	}

	result, err := h.service.SubmitDraft(c.Request.Context(), evaluationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, resp.Error(nil))
		return
	}

	c.JSON(http.StatusOK, resp.Success(result, "Draft submitted successfully"))
}

// @Summary Delete a draft evaluation
// @Description Delete a draft evaluation permanently.
// @Tags Evaluation
// @Produce json
// @Param id path string true "Draft Evaluation ID"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /evaluations/drafts/{id} [delete]
func (h *EvaluationHandler) DeleteDraft(c *gin.Context) {
	evaluationID := c.Param("id")
	if evaluationID == "" {
		c.JSON(http.StatusBadRequest, resp.Error(nil))
		return
	}

	err := h.service.DeleteDraft(c.Request.Context(), evaluationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}

	c.JSON(http.StatusOK, resp.Success(struct{}{}, "Draft deleted successfully"))
}
