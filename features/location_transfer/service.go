package locTransfer

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"
	"time"

	"go.uber.org/zap"
)

type locTransferService struct {
	logger *logger.Logger
	db     *db.Store
}

func NewLocationTransferService(db *db.Store, logger *logger.Logger) LocationTransferService {
	return &locTransferService{
		logger: logger,
		db:     db,
	}
}

func (s *locTransferService) RegisterLocationTransfer(ctx context.Context, req *RegisterLocationTransferRequest) (*RegisterLocationTransferResponse, error) {
	client, err := s.db.GetClientByID(ctx, req.ClientID)
	if err != nil {
		s.logger.Error(ctx, "RegisterLocationTransfer", "Failed to get client", zap.Error(err))
		return nil, ErrClientNotFound
	}

	result, err := s.db.CreateLocationTransfer(ctx, db.CreateLocationTransferParams{
		ID:                   nanoid.Generate(),
		ClientID:             client.ID,
		FromLocationID:       &client.AssignedLocationID,
		ToLocationID:         req.NewLocationID,
		NewCoordinatorID:     req.NewCoordinatorID,
		CurrentCoordinatorID: client.CoordinatorID,
	})
	if err != nil {
		s.logger.Error(ctx, "RegisterLocationTransfer", "Failed to create location transfer", zap.Error(err))
		return nil, ErrInternal
	}

	return &RegisterLocationTransferResponse{
		TransferID: result.ID,
	}, nil
}

func (s *locTransferService) ListLocationTransfers(ctx context.Context, req *ListLocationTransfersRequest) (*resp.PaginationResponse[ListLocationTransfersResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	transfers, err := s.db.ListLocationTransfers(ctx, db.ListLocationTransfersParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListLocationTransfers", "Failed to list location transfers", zap.Error(err))
		return nil, ErrInternal
	}

	listTransfersResponse := []ListLocationTransfersResponse{}
	totalCount := 0

	for _, transfer := range transfers {
		listTransfersResponse = append(listTransfersResponse, ListLocationTransfersResponse{
			ID:                          transfer.ID,
			ClientID:                    transfer.ClientID,
			FromLocationID:              transfer.FromLocationID,
			ToLocationID:                transfer.ToLocationID,
			CurrentCoordinatorID:        transfer.CurrentCoordinatorID,
			NewCoordinatorID:            transfer.NewCoordinatorID,
			TransferDate:                transfer.TransferDate.Time.Format(time.RFC3339),
			Reason:                      transfer.Reason,
			ClientFirstName:             transfer.ClientFirstName,
			ClientLastName:              transfer.ClientLastName,
			FromLocationName:            transfer.FromLocationName,
			ToLocationName:              transfer.ToLocationName,
			CurrentCoordinatorFirstName: transfer.CurrentCoordinatorFirstName,
			CurrentCoordinatorLastName:  transfer.CurrentCoordinatorLastName,
			NewCoordinatorFirstName:     transfer.NewCoordinatorFirstName,
			NewCoordinatorLastName:      transfer.NewCoordinatorLastName,
		})
		if totalCount == 0 {
			totalCount = int(transfer.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listTransfersResponse, totalCount, page, pageSize)
	return &result, nil
}



