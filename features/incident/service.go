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

func (s *incidentService) CreateIncident(
	ctx context.Context,
	req *CreateIncidentRequest,
) (CreateIncidentResponse, error) {
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

func (s *incidentService) ListIncidents(
	ctx context.Context,
	req *ListIncidentsRequest,
) (*resp.PaginationResponse[ListIncidentsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	var incidents []db.ListIncidentsRow
	var err error
	err = s.store.ExecTx(ctx, func(tx *db.Queries) error {
		incidents, err = tx.ListIncidents(ctx, db.ListIncidentsParams{
			Limit:  limit,
			Offset: offset,
			Search: req.Search,
		})
		if err != nil {
			s.logger.Error(ctx, "ListIncidents", "Failed to list incidents", zap.Error(err))
			return ErrInternal
		}
		return nil
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

func (s *incidentService) GetIncidentStats(
	ctx context.Context,
) (*GetIncidentStatsResponse, error) {
	var stats db.GetIncidentStatsRow
	var err error
	err = s.store.ExecTx(ctx, func(tx *db.Queries) error {
		stats, err = tx.GetIncidentStats(ctx)
		if err != nil {
			s.logger.Error(ctx, "GetIncidentStats", "Failed to get incident statistics", zap.Error(err))
			return ErrInternal
		}
		return nil
	})
	if err != nil {
		s.logger.Error(ctx, "GetIncidentStats", "Failed to get incident statistics", zap.Error(err))
		return nil, ErrInternal
	}

	return &GetIncidentStatsResponse{
		TotalCount: int(stats.TotalCount),
		CountsBySeverity: IncidentSeverityCountsDTO{
			Minor:    int(stats.MinorCount),
			Moderate: int(stats.ModerateCount),
			Severe:   int(stats.SevereCount),
		},
		CountsByStatus: IncidentStatusCountsDTO{
			Pending:            int(stats.PendingCount),
			UnderInvestigation: int(stats.UnderInvestigationCount),
			Completed:          int(stats.CompletedCount),
		},
		CountsByType: IncidentTypeCountsDTO{
			Aggression:       int(stats.AggressionCount),
			MedicalEmergency: int(stats.MedicalEmergencyCount),
			SafetyConcern:    int(stats.SafetyConcernCount),
			UnwantedBehavior: int(stats.UnwantedBehaviorCount),
			Other:            int(stats.OtherTypeCount),
		},
	}, nil
}

func (s *incidentService) GetIncident(
	ctx context.Context,
	id string,
) (*GetIncidentResponse, error) {
	var incident db.GetIncidentRow
	var err error
	err = s.store.ExecTx(ctx, func(tx *db.Queries) error {
		incident, err = tx.GetIncident(ctx, id)
		if err != nil {
			s.logger.Error(ctx, "GetIncident", "Failed to get incident", zap.Error(err))
			return ErrNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &GetIncidentResponse{
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
	}, nil
}

func (s *incidentService) UpdateIncident(
	ctx context.Context,
	id string,
	req *UpdateIncidentRequest,
) (*UpdateIncidentResponse, error) {
	var incidentDate pgtype.Date
	if req.IncidentDate != nil {
		incidentDate = pgtype.Date{Time: *req.IncidentDate, Valid: true}
	}

	var incidentTime pgtype.Time
	if req.IncidentTime != nil {
		incidentTime = util.TimeToPgtypeTime(*req.IncidentTime)
	}

	var incidentType db.NullIncidentTypeEnum
	if req.IncidentType != nil {
		incidentType = db.NullIncidentTypeEnum{
			IncidentTypeEnum: db.IncidentTypeEnum(*req.IncidentType),
			Valid:            true,
		}
	}

	var incidentSeverity db.NullIncidentSeverityEnum
	if req.IncidentSeverity != nil {
		incidentSeverity = db.NullIncidentSeverityEnum{
			IncidentSeverityEnum: db.IncidentSeverityEnum(*req.IncidentSeverity),
			Valid:                true,
		}
	}

	var status db.NullIncidentStatusEnum
	if req.Status != nil {
		status = db.NullIncidentStatusEnum{
			IncidentStatusEnum: db.IncidentStatusEnum(*req.Status),
			Valid:              true,
		}
	}

	err := s.store.ExecTx(ctx, func(tx *db.Queries) error {
		err := tx.UpdateIncident(ctx, db.UpdateIncidentParams{
			ID:                  id,
			IncidentDate:        incidentDate,
			IncidentTime:        incidentTime,
			IncidentType:        incidentType,
			IncidentSeverity:    incidentSeverity,
			LocationID:          req.LocationID,
			CoordinatorID:       req.CoordinatorID,
			IncidentDescription: req.IncidentDescription,
			ActionTaken:         req.ActionTaken,
			OtherParties:        req.OtherParties,
			Status:              status,
		})
		if err != nil {
			s.logger.Error(ctx, "UpdateIncident", "Failed to update incident", zap.Error(err))
			return ErrInternal
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateIncidentResponse{
		Success: true,
	}, nil
}

func (s *incidentService) DeleteIncident(
	ctx context.Context,
	id string,
) (*DeleteIncidentResponse, error) {
	err := s.store.ExecTx(ctx, func(tx *db.Queries) error {
		err := tx.SoftDeleteIncident(ctx, id)
		if err != nil {
			s.logger.Error(ctx, "DeleteIncident", "Failed to delete incident", zap.Error(err))
			return ErrInternal
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DeleteIncidentResponse{
		Success: true,
	}, nil
}
