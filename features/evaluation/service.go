package evaluation

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"context"
	"errors"

	"go.uber.org/zap"
)

type EvaluationService interface {
	CreateEvaluation(ctx context.Context, req *CreateEvaluationRequest) (*CreateEvaluationResponse, error)
	UpdateEvaluation(ctx context.Context, evaluationID string, req *UpdateEvaluationRequest) (*UpdateEvaluationResponse, error)
	GetEvaluationHistory(ctx context.Context, clientID string) ([]EvaluationHistoryItem, error)
	GetCriticalEvaluations(ctx context.Context) (*resp.PaginationResponse[UpcomingEvaluationDTO], error)
	GetScheduledEvaluations(ctx context.Context) (*resp.PaginationResponse[UpcomingEvaluationDTO], error)
	GetRecentEvaluations(ctx context.Context) (*resp.PaginationResponse[GlobalRecentEvaluationDTO], error)
	GetLastEvaluation(ctx context.Context, clientID string) (*LastEvaluationDTO, error)
	// Draft methods
	SaveDraft(ctx context.Context, req *SaveDraftRequest) (*SaveDraftResponse, error)
	GetDrafts(ctx context.Context) (*resp.PaginationResponse[DraftEvaluationListItemDTO], error)
	GetDraft(ctx context.Context, evaluationID string) (*DraftEvaluationDTO, error)
	SubmitDraft(ctx context.Context, evaluationID string) (*CreateEvaluationResponse, error)
	DeleteDraft(ctx context.Context, evaluationID string) error
}

type evaluationService struct {
	db     *db.Store
	logger logger.Logger
}

func NewEvaluationService(db *db.Store, logger logger.Logger) EvaluationService {
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

	// Determine evaluation status
	status := db.EvaluationStatusEnumSubmitted
	if req.IsDraft {
		status = db.EvaluationStatusEnumDraft
	}

	result, err := s.db.CreateEvaluationTx(ctx, db.CreateEvaluationTxParams{
		Evaluation: db.CreateClientEvaluationParams{
			ID:             evaluationID,
			ClientID:       req.ClientID,
			CoordinatorID:  req.CoordinatorID,
			EvaluationDate: util.StrToPgtypeDate(req.EvaluationDate),
			OverallNotes:   req.OverallNotes,
			Status:         status,
		},
		ProgressLogs:  progressLogs,
		IntervalWeeks: interval,
	})

	if err != nil {
		s.logger.Error(ctx, "CreateEvaluation", "Failed to create evaluation tx", zap.Error(err))
		return nil, err
	}

	response := &CreateEvaluationResponse{
		ID:      result.EvaluationID,
		IsDraft: req.IsDraft,
	}

	// Only set next evaluation date if not a draft
	if !req.IsDraft {
		response.NextEvaluationDate = &result.NextEvaluationDate.Time
	}

	return response, nil
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
			HasDraft:                row.HasDraft,
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
			HasDraft:                row.HasDraft,
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
			ClientID:             row.ClientID,
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

func (s *evaluationService) GetLastEvaluation(ctx context.Context, clientID string) (*LastEvaluationDTO, error) {
	rows, err := s.db.GetLastClientEvaluation(ctx, clientID)
	if err != nil {
		s.logger.Error(ctx, "GetLastEvaluation", "Failed to get last evaluation", zap.Error(err))
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil // No evaluation found for this client
	}

	// All rows have the same evaluation metadata, use the first row
	firstRow := rows[0]

	// Transform all rows into goal progress items
	goalProgress := util.Map(rows, func(row db.GetLastClientEvaluationRow) GoalProgressItemDTO {
		return GoalProgressItemDTO{
			GoalID:        row.GoalID,
			GoalTitle:     row.GoalTitle,
			Status:        string(row.Status),
			ProgressNotes: row.ProgressNotes,
		}
	})

	return &LastEvaluationDTO{
		EvaluationID:   firstRow.EvaluationID,
		EvaluationDate: firstRow.EvaluationDate.Time,
		OverallNotes:   firstRow.OverallNotes,
		Coordinator: CoordinatorInfoDTO{
			FirstName: firstRow.CoordinatorFirstName,
			LastName:  firstRow.CoordinatorLastName,
		},
		GoalProgress: goalProgress,
	}, nil
}

// SaveDraft creates or updates a draft evaluation
func (s *evaluationService) SaveDraft(ctx context.Context, req *SaveDraftRequest) (*SaveDraftResponse, error) {
	// Check if a draft already exists for this client
	existingDraft, err := s.db.GetDraftByClientId(ctx, req.ClientID)
	if err != nil && err.Error() != "no rows in result set" {
		s.logger.Error(ctx, "SaveDraft", "Failed to check for existing draft", zap.Error(err))
		return nil, err
	}

	var evaluationID string
	if existingDraft.ID != "" {
		// Update existing draft
		evaluationID = existingDraft.ID

		// Delete existing progress logs
		if err := s.db.DeleteGoalProgressLogsByEvaluationId(ctx, evaluationID); err != nil {
			s.logger.Error(ctx, "SaveDraft", "Failed to delete old progress logs", zap.Error(err))
			return nil, err
		}
	} else {
		// Create new draft
		evaluationID = nanoid.Generate()
		_, err := s.db.CreateClientEvaluation(ctx, db.CreateClientEvaluationParams{
			ID:             evaluationID,
			ClientID:       req.ClientID,
			CoordinatorID:  req.CoordinatorID,
			EvaluationDate: util.StrToPgtypeDate(req.EvaluationDate),
			OverallNotes:   req.OverallNotes,
			Status:         db.EvaluationStatusEnumDraft,
		})
		if err != nil {
			s.logger.Error(ctx, "SaveDraft", "Failed to create draft evaluation", zap.Error(err))
			return nil, err
		}
	}

	// Create/update progress logs
	for _, log := range req.ProgressLogs {
		if err := s.db.CreateGoalProgressLog(ctx, db.CreateGoalProgressLogParams{
			ID:            nanoid.Generate(),
			EvaluationID:  evaluationID,
			GoalID:        log.GoalID,
			Status:        db.GoalProgressStatus(log.Status),
			ProgressNotes: log.ProgressNotes,
		}); err != nil {
			s.logger.Error(ctx, "SaveDraft", "Failed to create progress log", zap.Error(err))
			return nil, err
		}
	}

	// Get the updated evaluation to return timestamp
	updatedEval, err := s.db.GetDraftByClientId(ctx, req.ClientID)
	if err != nil {
		s.logger.Error(ctx, "SaveDraft", "Failed to retrieve updated draft", zap.Error(err))
		return nil, err
	}

	return &SaveDraftResponse{
		ID:        updatedEval.ID,
		UpdatedAt: updatedEval.UpdatedAt.Time,
	}, nil
}

// GetDrafts retrieves all draft evaluations for a coordinator
func (s *evaluationService) GetDrafts(ctx context.Context) (*resp.PaginationResponse[DraftEvaluationListItemDTO], error) {
	coordinatorID := util.GetEmployeeID(ctx)
	if coordinatorID == "" {
		return nil, errors.New("user id is required")
	}

	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	rows, err := s.db.GetCoordinatorDrafts(ctx, db.GetCoordinatorDraftsParams{
		CoordinatorID: coordinatorID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		s.logger.Error(ctx, "GetDrafts", "Failed to get coordinator drafts", zap.Error(err))
		return nil, err
	}

	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}

	result := util.Map(rows, func(row db.GetCoordinatorDraftsRow) DraftEvaluationListItemDTO {
		return DraftEvaluationListItemDTO{
			EvaluationID:    row.EvaluationID,
			ClientID:        row.ClientID,
			ClientFirstName: row.ClientFirstName,
			ClientLastName:  row.ClientLastName,
			EvaluationDate:  row.EvaluationDate.Time,
			GoalsCount:      int(row.GoalsCount),
			UpdatedAt:       row.UpdatedAt.Time,
		}
	})

	pag := resp.PagResp(result, int(totalCount), int(page), int(pageSize))
	return &pag, nil
}

// GetDraft retrieves a specific draft evaluation with all progress logs
func (s *evaluationService) GetDraft(ctx context.Context, evaluationID string) (*DraftEvaluationDTO, error) {
	rows, err := s.db.GetDraftEvaluation(ctx, evaluationID)
	if err != nil {
		s.logger.Error(ctx, "GetDraft", "Failed to get draft evaluation", zap.Error(err))
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil // Draft not found
	}

	firstRow := rows[0]

	// Build goal progress list (filter out empty goals from LEFT JOIN)
	goalProgress := []GoalProgressItemDTO{}
	for _, row := range rows {
		if row.GoalID != nil {
			goalProgress = append(goalProgress, GoalProgressItemDTO{
				GoalID:        *row.GoalID,
				GoalTitle:     *row.GoalTitle,
				Status:        string(row.Status.GoalProgressStatus),
				ProgressNotes: row.ProgressNotes,
			})
		}
	}

	return &DraftEvaluationDTO{
		EvaluationID:         firstRow.EvaluationID,
		ClientID:             firstRow.ClientID,
		ClientFirstName:      firstRow.ClientFirstName,
		ClientLastName:       firstRow.ClientLastName,
		EvaluationDate:       firstRow.EvaluationDate.Time,
		OverallNotes:         firstRow.OverallNotes,
		CoordinatorFirstName: firstRow.CoordinatorFirstName,
		CoordinatorLastName:  firstRow.CoordinatorLastName,
		GoalProgress:         goalProgress,
		CreatedAt:            firstRow.CreatedAt.Time,
		UpdatedAt:            firstRow.UpdatedAt.Time,
	}, nil
}

// SubmitDraft converts a draft evaluation to submitted status
func (s *evaluationService) SubmitDraft(ctx context.Context, evaluationID string) (*CreateEvaluationResponse, error) {
	// Get the draft to retrieve client info
	draft, err := s.GetDraft(ctx, evaluationID)
	if err != nil {
		s.logger.Error(ctx, "SubmitDraft", "Failed to get draft for submission", zap.Error(err))
		return nil, err
	}
	if draft == nil {
		return nil, nil // Draft not found
	}

	// Get client to calculate next evaluation date
	client, err := s.db.GetClientByID(ctx, draft.ClientID)
	if err != nil {
		s.logger.Error(ctx, "SubmitDraft", "Failed to get client", zap.Error(err))
		return nil, err
	}

	// Submit the draft
	submittedEval, err := s.db.SubmitDraftEvaluation(ctx, evaluationID)
	if err != nil {
		s.logger.Error(ctx, "SubmitDraft", "Failed to submit draft", zap.Error(err))
		return nil, err
	}

	// Calculate and update next evaluation date
	interval := int32(5) // Default
	if client.EvaluationIntervalWeeks != nil {
		interval = *client.EvaluationIntervalWeeks
	}

	nextDate := submittedEval.EvaluationDate.Time.AddDate(0, 0, int(interval)*7)
	nextEvalDate := util.TimeToPgtypeDate(nextDate)

	if err := s.db.UpdateClientNextEvaluationDate(ctx, db.UpdateClientNextEvaluationDateParams{
		ID:                 client.ID,
		NextEvaluationDate: nextEvalDate,
	}); err != nil {
		s.logger.Error(ctx, "SubmitDraft", "Failed to update next evaluation date", zap.Error(err))
		return nil, err
	}

	return &CreateEvaluationResponse{
		ID:                 submittedEval.ID,
		NextEvaluationDate: &nextDate,
		IsDraft:            false,
	}, nil
}

// DeleteDraft deletes a draft evaluation
func (s *evaluationService) DeleteDraft(ctx context.Context, evaluationID string) error {
	if err := s.db.DeleteDraftEvaluation(ctx, evaluationID); err != nil {
		s.logger.Error(ctx, "DeleteDraft", "Failed to delete draft", zap.Error(err))
		return err
	}
	return nil
}

// UpdateEvaluation updates an existing submitted evaluation
func (s *evaluationService) UpdateEvaluation(ctx context.Context, evaluationID string, req *UpdateEvaluationRequest) (*UpdateEvaluationResponse, error) {
	// Build progress log params
	progressLogs := util.Map(req.ProgressLogs, func(log GoalProgressDTO) db.UpdateGoalProgressLogParams {
		return db.UpdateGoalProgressLogParams{
			EvaluationID:  evaluationID,
			GoalID:        log.GoalID,
			Status:        db.GoalProgressStatus(log.Status),
			ProgressNotes: log.ProgressNotes,
		}
	})

	// Execute update transaction
	result, err := s.db.UpdateEvaluationTx(ctx, db.UpdateEvaluationTxParams{
		EvaluationID:   evaluationID,
		EvaluationDate: util.StrToPgtypeDate(req.EvaluationDate),
		OverallNotes:   req.OverallNotes,
		ProgressLogs:   progressLogs,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateEvaluation", "Failed to update evaluation tx", zap.Error(err))
		return nil, err
	}

	return &UpdateEvaluationResponse{
		ID: result.EvaluationID,
	}, nil
}
