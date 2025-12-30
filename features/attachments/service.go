package attachments

import (
	"care-cordination/lib/bucket"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"context"
	"mime/multipart"

	"go.uber.org/zap"
)

type attachmentsService struct {
	db     *db.Store
	bucket bucket.ObjectStorage
	logger logger.Logger
}

func NewAttachmentsService(
	db *db.Store,
	bucket bucket.ObjectStorage,
	logger logger.Logger,
) AttachmentsService {
	return &attachmentsService{
		db:     db,
		bucket: bucket,
		logger: logger,
	}
}

func (s *attachmentsService) UploadAttachment(
	ctx context.Context,
	file *multipart.FileHeader,
) (*UploadAttachmentResponse, error) {
	id := nanoid.Generate()

	// Open the file
	src, err := file.Open()
	if err != nil {
		s.logger.Error(ctx, "UploadAttachment", "Failed to open file", zap.Error(err))
		return nil, ErrInvalidFile
	}
	defer src.Close()

	// Upload to object storage
	fileKey, err := s.bucket.UploadObject(ctx, id, src, file.Header.Get("Content-Type"))
	if err != nil {
		s.logger.Error(
			ctx,
			"UploadAttachment",
			"Failed to upload file to object storage",
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	// Save attachment metadata to database
	err = s.db.CreateAttachment(ctx, db.CreateAttachmentParams{
		ID:          id,
		Filekey:     fileKey,
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"UploadAttachment",
			"Failed to create attachment record",
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	return &UploadAttachmentResponse{
		ID: id,
	}, nil
}
