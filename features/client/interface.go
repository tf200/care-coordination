package client

import (
	"care-cordination/lib/resp"
	"context"
)

type ClientService interface {
	MoveClientToWaitingList(ctx context.Context, req *MoveClientToWaitingListRequest) (*MoveClientToWaitingListResponse, error)
	MoveClientInCare(ctx context.Context, clientID string, req *MoveClientInCareRequest) (*MoveClientInCareResponse, error)
	DischargeClient(ctx context.Context, clientID string, req *DischargeClientRequest) (*DischargeClientResponse, error)
	ListWaitingListClients(ctx context.Context, req *ListWaitingListClientsRequest) (*resp.PaginationResponse[ListWaitingListClientsResponse], error)
	ListInCareClients(ctx context.Context, req *ListInCareClientsRequest) (*resp.PaginationResponse[ListInCareClientsResponse], error)
}
