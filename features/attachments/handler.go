package attachments

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AttachmentsHandler struct {
	attachmentsService AttachmentsService
	mdw                *middleware.Middleware
}

func NewAttachmentsHandler(attachmentsService AttachmentsService, mdw *middleware.Middleware) *AttachmentsHandler {
	return &AttachmentsHandler{
		attachmentsService: attachmentsService,
		mdw:                mdw,
	}
}

func (h *AttachmentsHandler) SetupAttachmentsRoutes(router *gin.Engine) {
	attachments := router.Group("/attachments")

	attachments.POST("", h.mdw.AuthMiddleware(), h.UploadAttachment)
}

// @Summary Upload an attachment
// @Description Upload a file attachment
// @Tags Attachments
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} UploadAttachmentResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /attachments [post]
func (h *AttachmentsHandler) UploadAttachment(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.attachmentsService.UploadAttachment(ctx.Request.Context(), file)
	if err != nil {
		switch err {
		case ErrInvalidFile, ErrInvalidRequest:
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}
