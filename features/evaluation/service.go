package evaluation

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"context"

	"go.uber.org/zap"
)

type EvaluationService interface {
	CreateEvaluation(ctx context.Context, req *CreateEvaluationRequest) (*CreateEvaluationResponse, error)
	GetEvaluationHistory(ctx context.Context, clientID string) ([]EvaluationHistoryItem, error)
	GetCriticalEvaluations(ctx context.Context) (*resp.PaginationResponse[UpcomingEvaluationDTO], error)
	GetScheduledEvaluations(ctx context.Context) (*resp.PaginationResponse[UpcomingEvaluationDTO], error)
	GetRecentEvaluations(ctx context.Context) (*resp.PaginationResponse[GlobalRecentEvaluationDTO], error)
}

type evaluationService struct {
	db     *db.Store
	logger *logger.Logger
}

func NewEvaluationService(db *db.Store, logger *logger.Logger) EvaluationService {
	return &evaluationService{db: db, logger: logger}
}

func (s *evaluationService) CreateEvaluation(ctx context.Context, req *CreateEvaluationRequest) (*CreateEvaluationResponse, error) {
	client, err := s.db.GetClientByID(ctx, req.ClientID)
	if err != nil {
		s.logger.Error(ctx, "CreateEvaluation", "Failed to get client", zap.Error(err))
		return nil, err
	}

	evaluationID := nanoid.Generate()

	progressLogs := util.Map(req.ProgressLogs, func(log GoalProgressDTO) db.CreateGoalProgressLogParams {
		return db.CreateGoalProgressLogParams{
			ID:            nanoid.Generate(),
			EvaluationID:  evaluationID,
			GoalID:        log.GoalID,
			Status:        db.GoalProgressStatus(log.Status),
			ProgressNotes: log.ProgressNotes,
		}
	})

	interval := int32(5) // Default
	if client.EvaluationIntervalWeeks != nil {
		interval = *client.EvaluationIntervalWeeks
	}

	result, err := s.db.CreateEvaluationTx(ctx, db.CreateEvaluationTxParams{
		Evaluation: db.CreateClientEvaluationParams{
			ID:             evaluationID,
			ClientID:       req.ClientID,
			CoordinatorID:  req.CoordinatorID,
			EvaluationDate: util.StrToPgtypeDate(req.EvaluationDate),
			OverallNotes:   req.OverallNotes,
		},
		ProgressLogs:  progressLogs,
		IntervalWeeks: interval,
	})

	if err != nil {
		s.logger.Error(ctx, "CreateEvaluation", "Failed to create evaluation tx", zap.Error(err))
		return nil, err
	}

	return &CreateEvaluationResponse{
		ID:                 result.EvaluationID,
		NextEvaluationDate: result.NextEvaluationDate.Time,
	}, nil
}

func (s *evaluationService) GetEvaluationHistory(ctx context.Context, clientID string) ([]EvaluationHistoryItem, error) {
	rows, err := s.db.GetClientEvaluationHistory(ctx, clientID)
	if err != nil {
		s.logger.Error(ctx, "GetEvaluationHistory", "Failed to get history", zap.Error(err))
		return nil, err
	}

	return util.Map(rows, func(row db.GetClientEvaluationHistoryRow) EvaluationHistoryItem {
		return EvaluationHistoryItem{
			EvaluationID:         row.EvaluationID,
			EvaluationDate:       row.EvaluationDate.Time,
			OverallNotes:         row.OverallNotes,
			GoalID:               row.GoalID,
			GoalTitle:            row.GoalTitle,
			Status:               string(row.Status),
			ProgressNotes:        row.ProgressNotes,
			CoordinatorFirstName: row.CoordinatorFirstName,
			CoordinatorLastName:  row.CoordinatorLastName,
		}
	}), nil
}

func (s *evaluationService) GetCriticalEvaluations(ctx context.Context) (*resp.PaginationResponse[UpcomingEvaluationDTO], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	rows, err := s.db.GetCriticalEvaluations(ctx, db.GetCriticalEvaluationsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.logger.Error(ctx, "GetCriticalEvaluations", "Failed to get critical evaluations", zap.Error(err))
		return nil, err
	}

	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}

	result := util.Map(rows, func(row db.GetCriticalEvaluationsRow) UpcomingEvaluationDTO {
		return UpcomingEvaluationDTO{
			ID:                      row.ID,
			FirstName:               row.FirstName,
			LastName:                row.LastName,
			NextEvaluationDate:      row.NextEvaluationDate.Time,
			EvaluationIntervalWeeks: int(util.PointerInt32ToIntValue(row.EvaluationIntervalWeeks)),
			LocationName:            row.LocationName,
			CoordinatorFirstName:    row.CoordinatorFirstName,
			CoordinatorLastName:     row.CoordinatorLastName,
		}
	})

	pag := resp.PagResp(result, int(totalCount), int(page), int(pageSize))
	return &pag, nil
}

func (s *evaluationService) GetScheduledEvaluations(ctx context.Context) (*resp.PaginationResponse[UpcomingEvaluationDTO], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	rows, err := s.db.GetScheduledEvaluations(ctx, db.GetScheduledEvaluationsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.logger.Error(ctx, "GetScheduledEvaluations", "Failed to get scheduled evaluations", zap.Error(err))
		return nil, err
	}

	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}

	result := util.Map(rows, func(row db.GetScheduledEvaluationsRow) UpcomingEvaluationDTO {
		return UpcomingEvaluationDTO{
			ID:                      row.ID,
			FirstName:               row.FirstName,
			LastName:                row.LastName,
			NextEvaluationDate:      row.NextEvaluationDate.Time,
			EvaluationIntervalWeeks: int(util.PointerInt32ToIntValue(row.EvaluationIntervalWeeks)),
			LocationName:            row.LocationName,
			CoordinatorFirstName:    row.CoordinatorFirstName,
			CoordinatorLastName:     row.CoordinatorLastName,
		}
	})

	pag := resp.PagResp(result, int(totalCount), int(page), int(pageSize))
	return &pag, nil
}

func (s *evaluationService) GetRecentEvaluations(ctx context.Context) (*resp.PaginationResponse[GlobalRecentEvaluationDTO], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	rows, err := s.db.GetRecentEvaluationsGlobal(ctx, db.GetRecentEvaluationsGlobalParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.logger.Error(ctx, "GetRecentEvaluations", "Failed to get recent history", zap.Error(err))
		return nil, err
	}

	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}

	result := util.Map(rows, func(row db.GetRecentEvaluationsGlobalRow) GlobalRecentEvaluationDTO {
		return GlobalRecentEvaluationDTO{
			EvaluationID:         row.EvaluationID,
			EvaluationDate:       row.EvaluationDate.Time,
			ClientFirstName:      row.ClientFirstName,
			ClientLastName:       row.ClientLastName,
			CoordinatorFirstName: row.CoordinatorFirstName,
			CoordinatorLastName:  row.CoordinatorLastName,
			TotalGoals:           int(row.TotalGoals),
			GoalsAchieved:        int(row.GoalsAchieved),
		}
	})

	pag := resp.PagResp(result, int(totalCount), int(page), int(pageSize))
	return &pag, nil
}
