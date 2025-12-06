package client

import (
	"care-cordination/lib/resp"
	"context"
)

type ClientService interface {
	MoveClientToWaitingList(ctx context.Context, req *MoveClientToWaitingListRequest) (*MoveClientToWaitingListResponse, error)
	MoveClientInCare(ctx context.Context, clientID string, req *MoveClientInCareRequest) (*MoveClientInCareResponse, error)
	StartDischarge(ctx context.Context, clientID string, req *StartDischargeRequest) (*StartDischargeResponse, error)
	CompleteDischarge(ctx context.Context, clientID string, req *CompleteDischargeRequest) (*CompleteDischargeResponse, error)
	ListWaitingListClients(ctx context.Context, req *ListWaitingListClientsRequest) (*resp.PaginationResponse[ListWaitingListClientsResponse], error)
	ListInCareClients(ctx context.Context, req *ListInCareClientsRequest) (*resp.PaginationResponse[ListInCareClientsResponse], error)
	ListDischargedClients(ctx context.Context, req *ListDischargedClientsRequest) (*resp.PaginationResponse[ListDischargedClientsResponse], error)
}
