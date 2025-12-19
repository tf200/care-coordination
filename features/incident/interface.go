package incident

import (
	"care-cordination/lib/resp"
	"context"
)

type IncidentService interface {
	CreateIncident(ctx context.Context, req *CreateIncidentRequest) (CreateIncidentResponse, error)
	ListIncidents(
		ctx context.Context,
		req *ListIncidentsRequest,
	) (*resp.PaginationResponse[ListIncidentsResponse], error)
}
