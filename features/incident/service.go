package incident

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type incidentService struct {
	store  *db.Store
	logger *logger.Logger
}

func NewIncidentService(store *db.Store, logger *logger.Logger) IncidentService {
	return &incidentService{
		store:  store,
		logger: logger,
	}
}

func (s *incidentService) CreateIncident(ctx context.Context, req *CreateIncidentRequest) (CreateIncidentResponse, error) {
	id := nanoid.Generate()

	// Handle optional other_parties
	var otherParties *string
	if req.OtherParties != "" {
		otherParties = &req.OtherParties
	}

	err := s.store.CreateIncident(ctx, db.CreateIncidentParams{
		ClientID:            req.ClientID,
		IncidentDate:        pgtype.Date{Time: req.IncidentDate, Valid: true},
		IncidentTime:        util.TimeToPgtypeTime(req.IncidentTime),
		IncidentType:        db.IncidentTypeEnum(req.IncidentType),
		IncidentSeverity:    db.IncidentSeverityEnum(req.IncidentSeverity),
		LocationID:          req.LocationID,
		CoordinatorID:       req.CoordinatorID,
		IncidentDescription: req.IncidentDescription,
		ActionTaken:         req.ActionTaken,
		OtherParties:        otherParties,
		Status:              db.IncidentStatusEnum(req.Status),
	})
	if err != nil {
		s.logger.Error(ctx, "CreateIncident", "Failed to create incident", zap.Error(err))
		return CreateIncidentResponse{}, ErrInternal
	}

	return CreateIncidentResponse{
		ID: id,
	}, nil
}

func (s *incidentService) ListIncidents(ctx context.Context, req *ListIncidentsRequest) (*resp.PaginationResponse[ListIncidentsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	incidents, err := s.store.ListIncidents(ctx, db.ListIncidentsParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListIncidents", "Failed to list incidents", zap.Error(err))
		return nil, ErrInternal
	}

	listIncidentsResponse := []ListIncidentsResponse{}
	totalCount := 0

	for _, incident := range incidents {
		listIncidentsResponse = append(listIncidentsResponse, ListIncidentsResponse{
			ID:                   incident.ID,
			ClientID:             incident.ClientID,
			ClientFirstName:      incident.ClientFirstName,
			ClientLastName:       incident.ClientLastName,
			IncidentDate:         incident.IncidentDate.Time,
			IncidentTime:         util.PgtypeTimeToString(incident.IncidentTime),
			IncidentType:         string(incident.IncidentType),
			IncidentSeverity:     string(incident.IncidentSeverity),
			LocationID:           incident.LocationID,
			LocationName:         incident.LocationName,
			CoordinatorID:        incident.CoordinatorID,
			CoordinatorFirstName: incident.CoordinatorFirstName,
			CoordinatorLastName:  incident.CoordinatorLastName,
			IncidentDescription:  incident.IncidentDescription,
			ActionTaken:          incident.ActionTaken,
			OtherParties:         incident.OtherParties,
			Status:               string(incident.Status),
			CreatedAt:            incident.CreatedAt.Time,
		})
		if totalCount == 0 {
			totalCount = int(incident.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listIncidentsResponse, totalCount, page, pageSize)
	return &result, nil
}
