package client

import "context"

type ClientService interface {
	MoveClientToWaitingList(ctx context.Context, req *MoveClientToWaitingListRequest) (*MoveClientToWaitingListResponse, error)
	MoveClientInCare(ctx context.Context, clientID string, req *MoveClientInCareRequest) (*MoveClientInCareResponse, error)
	DischargeClient(ctx context.Context, clientID string, req *DischargeClientRequest) (*DischargeClientResponse, error)
}
