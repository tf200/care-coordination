package incident

import (
	"care-cordination/lib/resp"
	"context"
)

type IncidentService interface {
	CreateIncident(ctx context.Context, req *CreateIncidentRequest) (CreateIncidentResponse, error)
	GetIncident(ctx context.Context, id string) (*GetIncidentResponse, error)
	UpdateIncident(ctx context.Context, id string, req *UpdateIncidentRequest) (*UpdateIncidentResponse, error)
	DeleteIncident(ctx context.Context, id string) (*DeleteIncidentResponse, error)
	ListIncidents(
		ctx context.Context,
		req *ListIncidentsRequest,
	) (*resp.PaginationResponse[ListIncidentsResponse], error)

	GetIncidentStats(ctx context.Context) (*GetIncidentStatsResponse, error)
}
