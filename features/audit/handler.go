package audit

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service AuditService
	mdw     *middleware.Middleware
}

func NewAuditHandler(service AuditService, mdw *middleware.Middleware) *AuditHandler {
	return &AuditHandler{
		service: service,
		mdw:     mdw,
	}
}

func (h *AuditHandler) SetupAuditRoutes(router *gin.Engine) {
	audit := router.Group("/audit")
	audit.Use(h.mdw.AuthMdw())
	// Only admins should access audit logs
	audit.Use(h.mdw.RequirePermission("admin", "manage"))

	audit.GET("/logs", h.mdw.PaginationMdw(), h.ListAuditLogs)
	audit.GET("/logs/:id", h.GetAuditLog)
	audit.GET("/stats", h.GetAuditStats)
	audit.GET("/verify", h.VerifyFullChain)
	audit.GET("/verify/range", h.VerifyChainRange)
}

// @Summary List audit logs
// @Description List all audit logs with optional filters (admin only)
// @Tags Audit
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param resource_type query string false "Filter by resource type (client, intake_form, etc)"
// @Param resource_id query string false "Filter by resource ID"
// @Param action query string false "Filter by action (read, create, update, delete)"
// @Param start_date query string false "Filter from date (RFC3339 format)"
// @Param end_date query string false "Filter to date (RFC3339 format)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]AuditLogResponse]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 403 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /audit/logs [get]
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	var req ListAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	limit, offset, page, pageSize := middleware.GetPaginationParams(c)

	logs, totalCount, err := h.service.ListAuditLogs(c, &req, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
		return
	}

	result := resp.PagRespWithParams(logs, int(totalCount), page, pageSize)
	c.JSON(http.StatusOK, resp.Success(result, "Audit logs retrieved successfully"))
}

// @Summary Get audit log by ID
// @Description Get a specific audit log entry by ID (admin only)
// @Tags Audit
// @Produce json
// @Param id path string true "Audit Log ID"
// @Success 200 {object} resp.SuccessResponse[AuditLogResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 403 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /audit/logs/{id} [get]
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	log, err := h.service.GetAuditLogByID(c, id)
	if err != nil {
		switch err {
		case ErrAuditLogNotFound:
			c.JSON(http.StatusNotFound, resp.Error(err))
		default:
			c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
		}
		return
	}

	c.JSON(http.StatusOK, resp.Success(log, "Audit log retrieved successfully"))
}

// @Summary Get audit log statistics
// @Description Get statistics for audit logs in the last 24 hours (admin only)
// @Tags Audit
// @Produce json
// @Success 200 {object} resp.SuccessResponse[AuditLogStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 403 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /audit/stats [get]
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	stats, err := h.service.GetAuditStats(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, resp.Success(stats, "Audit statistics retrieved successfully"))
}

// @Summary Verify entire audit log chain
// @Description Verify the integrity of the entire audit log hash chain (admin only)
// @Tags Audit
// @Produce json
// @Success 200 {object} resp.SuccessResponse[ChainVerificationResult]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 403 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /audit/verify [get]
func (h *AuditHandler) VerifyFullChain(c *gin.Context) {
	result, err := h.service.VerifyFullChain(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
		return
	}

	message := "Audit log chain verification successful: chain is intact"
	if !result.IsValid {
		message = "WARNING: Audit log chain verification FAILED - tampering detected!"
	}

	c.JSON(http.StatusOK, resp.Success(result, message))
}

// @Summary Verify audit log chain range
// @Description Verify the integrity of a range of audit logs (admin only)
// @Tags Audit
// @Produce json
// @Param start_sequence query int true "Start sequence number"
// @Param end_sequence query int true "End sequence number"
// @Success 200 {object} resp.SuccessResponse[ChainVerificationResult]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 403 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /audit/verify/range [get]
func (h *AuditHandler) VerifyChainRange(c *gin.Context) {
	var req VerifyChainRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.service.VerifyChain(c, req.StartSequence, req.EndSequence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
		return
	}

	message := "Audit log chain verification successful: chain is intact"
	if !result.IsValid {
		message = "WARNING: Audit log chain verification FAILED - tampering detected!"
	}

	c.JSON(http.StatusOK, resp.Success(result, message))
}
