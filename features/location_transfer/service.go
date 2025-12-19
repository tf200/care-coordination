package locTransfer

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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

func (s *locTransferService) RegisterLocationTransfer(
	ctx context.Context,
	req *RegisterLocationTransferRequest,
) (*RegisterLocationTransferResponse, error) {
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
		s.logger.Error(
			ctx,
			"RegisterLocationTransfer",
			"Failed to create location transfer",
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	return &RegisterLocationTransferResponse{
		TransferID: result.ID,
	}, nil
}

func (s *locTransferService) ListLocationTransfers(
	ctx context.Context,
	req *ListLocationTransfersRequest,
) (*resp.PaginationResponse[ListLocationTransfersResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	transfers, err := s.db.ListLocationTransfers(ctx, db.ListLocationTransfersParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"ListLocationTransfers",
			"Failed to list location transfers",
			zap.Error(err),
		)
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
			Status:                      string(transfer.Status),
			RejectionReason:             transfer.RejectionReason,
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

func (s *locTransferService) GetLocationTransferByID(
	ctx context.Context,
	transferID string,
) (*ListLocationTransfersResponse, error) {
	transfer, err := s.db.GetLocationTransferByID(ctx, transferID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTransferNotFound
		}
		s.logger.Error(ctx, "GetLocationTransferByID", "Failed to get transfer", zap.Error(err))
		return nil, ErrInternal
	}

	return &ListLocationTransfersResponse{
		ID:                          transfer.ID,
		ClientID:                    transfer.ClientID,
		FromLocationID:              transfer.FromLocationID,
		ToLocationID:                transfer.ToLocationID,
		CurrentCoordinatorID:        transfer.CurrentCoordinatorID,
		NewCoordinatorID:            transfer.NewCoordinatorID,
		TransferDate:                transfer.TransferDate.Time.Format(time.RFC3339),
		Reason:                      transfer.Reason,
		Status:                      string(transfer.Status),
		RejectionReason:             transfer.RejectionReason,
		ClientFirstName:             transfer.ClientFirstName,
		ClientLastName:              transfer.ClientLastName,
		FromLocationName:            transfer.FromLocationName,
		ToLocationName:              transfer.ToLocationName,
		CurrentCoordinatorFirstName: transfer.CurrentCoordinatorFirstName,
		CurrentCoordinatorLastName:  transfer.CurrentCoordinatorLastName,
		NewCoordinatorFirstName:     transfer.NewCoordinatorFirstName,
		NewCoordinatorLastName:      transfer.NewCoordinatorLastName,
	}, nil
}

func (s *locTransferService) ConfirmLocationTransfer(
	ctx context.Context,
	transferID string,
) error {
	// First, get the transfer to check status and get details
	transfer, err := s.db.GetLocationTransferByID(ctx, transferID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrTransferNotFound
		}
		s.logger.Error(ctx, "ConfirmLocationTransfer", "Failed to get transfer", zap.Error(err))
		return ErrInternal
	}

	// Check if already processed
	if transfer.Status != db.LocationTransferStatusEnumPending {
		return ErrTransferAlreadyProcessed
	}

	// Execute all updates in a transaction
	err = s.db.ExecTx(ctx, func(q *db.Queries) error {
		// 1. Confirm the transfer
		if err := q.ConfirmLocationTransfer(ctx, transferID); err != nil {
			return err
		}

		// 2. Update client's location and coordinator
		if _, err := q.UpdateClient(ctx, db.UpdateClientParams{
			ID:                 transfer.ClientID,
			AssignedLocationID: &transfer.ToLocationID,
			CoordinatorID:      &transfer.NewCoordinatorID,
		}); err != nil {
			return err
		}

		// 3. Decrement old location's occupied count (if from_location exists)
		if transfer.FromLocationID != nil {
			if err := q.DecrementLocationOccupied(ctx, *transfer.FromLocationID); err != nil {
				return err
			}
		}

		// 4. Increment new location's occupied count
		if err := q.IncrementLocationOccupied(ctx, transfer.ToLocationID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "ConfirmLocationTransfer", "Transaction failed", zap.Error(err))
		return ErrInternal
	}

	return nil
}

func (s *locTransferService) RefuseLocationTransfer(
	ctx context.Context,
	transferID string,
	req *RefuseLocationTransferRequest,
) error {
	// First, get the transfer to check status
	transfer, err := s.db.GetLocationTransferByID(ctx, transferID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrTransferNotFound
		}
		s.logger.Error(ctx, "RefuseLocationTransfer", "Failed to get transfer", zap.Error(err))
		return ErrInternal
	}

	// Check if already processed
	if transfer.Status != db.LocationTransferStatusEnumPending {
		return ErrTransferAlreadyProcessed
	}

	err = s.db.RefuseLocationTransfer(ctx, db.RefuseLocationTransferParams{
		ID:              transferID,
		RejectionReason: &req.Reason,
	})
	if err != nil {
		s.logger.Error(ctx, "RefuseLocationTransfer", "Failed to refuse transfer", zap.Error(err))
		return ErrInternal
	}

	return nil
}

func (s *locTransferService) UpdateLocationTransfer(
	ctx context.Context,
	transferID string,
	req *UpdateLocationTransferRequest,
) error {
	// First, get the transfer to check status
	transfer, err := s.db.GetLocationTransferByID(ctx, transferID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrTransferNotFound
		}
		s.logger.Error(ctx, "UpdateLocationTransfer", "Failed to get transfer", zap.Error(err))
		return ErrInternal
	}

	// Check if already processed
	if transfer.Status != db.LocationTransferStatusEnumPending {
		return ErrTransferAlreadyProcessed
	}

	err = s.db.UpdateLocationTransfer(ctx, db.UpdateLocationTransferParams{
		ID:               transferID,
		ToLocationID:     req.NewLocationID,
		NewCoordinatorID: req.NewCoordinatorID,
		Reason:           req.Reason,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateLocationTransfer", "Failed to update transfer", zap.Error(err))
		return ErrInternal
	}

	return nil
}
