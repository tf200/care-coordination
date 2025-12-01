package attachments

import (
	"context"
	"mime/multipart"
)

type AttachmentsService interface {
	UploadAttachment(ctx context.Context, file *multipart.FileHeader) (*UploadAttachmentResponse, error)
}
