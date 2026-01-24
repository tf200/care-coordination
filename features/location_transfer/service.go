package locTransfer

import (
	"care-cordination/lib/middleware"
	"care-cordination/features/notification"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type locTransferService struct {
	logger              logger.Logger
	db                  *db.Store
	notificationService notification.NotificationService
}

func NewLocationTransferService(
	db *db.Store,
	logger logger.Logger,
	notificationService notification.NotificationService,
) LocationTransferService {
	return &locTransferService{
		logger:              logger,
		db:                  db,
		notificationService: notificationService,
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
	util.SetClientID(ctx, client.ID)

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

	util.SetClientID(ctx, transfer.ClientID)

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
	util.SetClientID(ctx, transfer.ClientID)

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

	// Trigger: Notify both coordinators about approved transfer
	if s.notificationService != nil {
		resourceType := notification.ResourceTypeLocationTransfer
		resourceID := transferID

		// Get user IDs for both coordinators
		currentCoordUserID := s.getEmployeeUserID(ctx, transfer.CurrentCoordinatorID)
		newCoordUserID := s.getEmployeeUserID(ctx, transfer.NewCoordinatorID)

		userIDs := []string{}
		if currentCoordUserID != "" {
			userIDs = append(userIDs, currentCoordUserID)
		}
		if newCoordUserID != "" && newCoordUserID != currentCoordUserID {
			userIDs = append(userIDs, newCoordUserID)
		}

		if len(userIDs) > 0 {
			toLocationName := ""
			if transfer.ToLocationName != nil {
				toLocationName = *transfer.ToLocationName
			}
			s.notificationService.EnqueueForUsers(userIDs, &notification.CreateNotificationRequest{
				Type:         notification.TypeLocationTransferApproved,
				Priority:     notification.PriorityNormal,
				Title:        "Location Transfer Approved",
				Message:      fmt.Sprintf("Client %s %s transferred to %s", transfer.ClientFirstName, transfer.ClientLastName, toLocationName),
				ResourceType: &resourceType,
				ResourceID:   &resourceID,
			})
		}
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

	util.SetClientID(ctx, transfer.ClientID)

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

	// Trigger: Notify requesting coordinator about rejection
	if s.notificationService != nil {
		resourceType := notification.ResourceTypeLocationTransfer
		resourceID := transferID

		currentCoordUserID := s.getEmployeeUserID(ctx, transfer.CurrentCoordinatorID)
		if currentCoordUserID != "" {
			s.notificationService.Enqueue(&notification.CreateNotificationRequest{
				UserID:       currentCoordUserID,
				Type:         notification.TypeLocationTransferRejected,
				Priority:     notification.PriorityNormal,
				Title:        "Location Transfer Rejected",
				Message:      fmt.Sprintf("Transfer request for %s %s was rejected: %s", transfer.ClientFirstName, transfer.ClientLastName, req.Reason),
				ResourceType: &resourceType,
				ResourceID:   &resourceID,
			})
		}
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

	util.SetClientID(ctx, transfer.ClientID)

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

func (s *locTransferService) GetLocationTransferStats(
	ctx context.Context,
) (*GetLocationTransferStatsResponse, error) {
	stats, err := s.db.GetLocationTransferStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetLocationTransferStats", "Failed to get transfer statistics", zap.Error(err))
		return nil, ErrInternal
	}

	// Convert approval_rate from int32 to float64
	approvalRate := float64(stats.ApprovalRate)

	return &GetLocationTransferStatsResponse{
		TotalCount:   int(stats.TotalCount),
		PendingCount: int(stats.PendingCount),
		ApprovalRate: approvalRate,
		CountsByStatus: TransferStatusCounts{
			Pending:  int(stats.PendingCount),
			Approved: int(stats.ApprovedCount),
			Rejected: int(stats.RejectedCount),
		},
	}, nil
}

// getEmployeeUserID looks up the user ID for an employee
func (s *locTransferService) getEmployeeUserID(ctx context.Context, employeeID string) string {
	employee, err := s.db.GetEmployeeByID(ctx, employeeID)
	if err != nil {
		return ""
	}
	return employee.UserID
}
