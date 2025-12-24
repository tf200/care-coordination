package locTransfer

import (
	"care-cordination/lib/resp"
	"context"
)

type LocationTransferService interface {
	RegisterLocationTransfer(
		ctx context.Context,
		req *RegisterLocationTransferRequest,
	) (*RegisterLocationTransferResponse, error)
	ListLocationTransfers(
		ctx context.Context,
		req *ListLocationTransfersRequest,
	) (*resp.PaginationResponse[ListLocationTransfersResponse], error)
	GetLocationTransferByID(
		ctx context.Context,
		transferID string,
	) (*ListLocationTransfersResponse, error)
	ConfirmLocationTransfer(
		ctx context.Context,
		transferID string,
	) error
	RefuseLocationTransfer(
		ctx context.Context,
		transferID string,
		req *RefuseLocationTransferRequest,
	) error
	UpdateLocationTransfer(
		ctx context.Context,
		transferID string,
		req *UpdateLocationTransferRequest,
	) error

	GetLocationTransferStats(ctx context.Context) (*GetLocationTransferStatsResponse, error)
}
