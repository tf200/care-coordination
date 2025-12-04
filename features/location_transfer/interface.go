package locTransfer

import (
	"care-cordination/lib/resp"
	"context"
)

type LocationTransferService interface {
	RegisterLocationTransfer(ctx context.Context, req *RegisterLocationTransferRequest) (*RegisterLocationTransferResponse, error)
	ListLocationTransfers(ctx context.Context, req *ListLocationTransfersRequest) (*resp.PaginationResponse[ListLocationTransfersResponse], error)
}
