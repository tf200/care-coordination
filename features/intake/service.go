package intake

import (
	"care-cordination/lib/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"context"

	"go.uber.org/zap"
)

type intakeService struct {
	db     *db.Store
	logger logger.Logger
}

func NewIntakeService(db *db.Store, logger logger.Logger) IntakeService {
	return &intakeService{
		db:     db,
		logger: logger,
	}
}

func (s *intakeService) CreateIntakeForm(
	ctx context.Context,
	req *CreateIntakeFormRequest,
) (*CreateIntakeFormResponse, error) {
	id := nanoid.Generate()
	_, err := s.db.CreateIntakeFormTx(ctx, db.CreateIntakeFormTxParams{
		IntakeForm: db.CreateIntakeFormParams{
			ID:                      id,
			RegistrationFormID:      req.RegistrationFormID,
			IntakeDate:              util.StrToPgtypeDate(req.IntakeDate),
			IntakeTime:              util.StrToPgtypeTime(req.IntakeTime),
			LocationID:              req.LocationID,
			CoordinatorID:           req.CoordinatorID,
			FamilySituation:         req.FamilySituation,
			MainProvider:            req.MainProvider,
			Limitations:             req.Limitations,
			FocusAreas:              req.FocusAreas,
			Notes:                   req.Notes,
			EvaluationIntervalWeeks: util.IntToPointerInt32(req.EvaluationInterval),
		},
		RegistrationFormID: req.RegistrationFormID,
		RegistrationFormStatus: db.NullRegistrationStatusEnum{
			RegistrationStatusEnum: db.RegistrationStatusEnumInReview,
			Valid:                  true,
		},
		Goals: util.Map(req.Goals, func(g GoalItem) db.CreateClientGoalParams {
			return db.CreateClientGoalParams{
				ID:           nanoid.Generate(),
				IntakeFormID: id,
				Title:        g.Title,
				Description:  g.Description,
			}
		}),
	})
	if err != nil {
		s.logger.Error(ctx, "CreateIntakeForm", "Failed to create intake form", zap.Error(err))
		return nil, ErrInternal
	}
	return &CreateIntakeFormResponse{
		ID: id,
	}, nil
}

func (s *intakeService) ListIntakeForms(
	ctx context.Context,
	req *ListIntakeFormsRequest,
) (*resp.PaginationResponse[ListIntakeFormsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	var intakeForms []db.ListIntakeFormsRow
	err := s.db.ExecTx(ctx, func(q *db.Queries) error {
		var err error
		intakeForms, err = q.ListIntakeForms(ctx, db.ListIntakeFormsParams{
			Limit:   limit,
			Offset:  offset,
			Column3: *req.Search,
		})
		return err
	})
	if err != nil {
		s.logger.Error(ctx, "ListIntakeForms", "Failed to list intake forms", zap.Error(err))
		return nil, ErrInternal
	}
	listIntakeFormsResponse := []ListIntakeFormsResponse{}
	for _, intakeForm := range intakeForms {
		var careType *string
		if intakeForm.CareType.Valid {
			ct := string(intakeForm.CareType.CareTypeEnum)
			careType = &ct
		}
		listIntakeFormsResponse = append(listIntakeFormsResponse, ListIntakeFormsResponse{
			ID:                   intakeForm.ID,
			RegistrationFormID:   intakeForm.RegistrationFormID,
			IntakeDate:           intakeForm.IntakeDate.Time,
			IntakeTime:           util.PgtypeTimeToString(intakeForm.IntakeTime),
			LocationID:           intakeForm.LocationID,
			CoordinatorID:        intakeForm.CoordinatorID,
			MainProvider:         intakeForm.MainProvider,
			ClientFirstName:      intakeForm.FirstName,
			ClientLastName:       intakeForm.LastName,
			ClientBSN:            intakeForm.Bsn,
			CareType:             careType,
			OrganizationName:     intakeForm.OrgName,
			LocationName:         intakeForm.LocationName,
			CoordinatorFirstName: intakeForm.CoordinatorFirstName,
			CoordinatorLastName:  intakeForm.CoordinatorLastName,
			Status:               string(intakeForm.Status),
		})
	}
	totalCount := 0
	if len(intakeForms) > 0 {
		totalCount = int(intakeForms[0].TotalCount)
	}
	// Use page and pageSize (not offset and limit) for correct pagination metadata
	result := resp.PagRespWithParams(listIntakeFormsResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *intakeService) GetIntakeForm(
	ctx context.Context,
	id string,
) (*GetIntakeFormResponse, error) {
	intakeForm, err := s.db.GetIntakeFormWithDetails(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "GetIntakeForm", "Failed to get intake form", zap.Error(err))
		return nil, ErrInternal
	}

	intakeGoals, err := s.db.ListGoalsByIntakeID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "GetIntakeForm", "Failed to list goals", zap.Error(err))
		return nil, ErrInternal
	}

	var careType *string
	if intakeForm.CareType.Valid {
		ct := string(intakeForm.CareType.CareTypeEnum)
		careType = &ct
	}

	return &GetIntakeFormResponse{
		ID:                   intakeForm.ID,
		RegistrationFormID:   intakeForm.RegistrationFormID,
		IntakeDate:           intakeForm.IntakeDate.Time,
		IntakeTime:           util.PgtypeTimeToString(intakeForm.IntakeTime),
		LocationID:           intakeForm.LocationID,
		CoordinatorID:        intakeForm.CoordinatorID,
		FamilySituation:      intakeForm.FamilySituation,
		MainProvider:         intakeForm.MainProvider,
		Limitations:          intakeForm.Limitations,
		FocusAreas:           intakeForm.FocusAreas,
		Notes:                intakeForm.Notes,
		EvaluationInterval:   util.PointerInt32ToIntValue(intakeForm.EvaluationIntervalWeeks),
		Status:               string(intakeForm.Status),
		ClientFirstName:      intakeForm.ClientFirstName,
		ClientLastName:       intakeForm.ClientLastName,
		ClientBSN:            intakeForm.ClientBsn,
		CareType:             careType,
		OrganizationName:     intakeForm.OrgName,
		LocationName:         intakeForm.LocationName,
		CoordinatorFirstName: intakeForm.CoordinatorFirstName,
		CoordinatorLastName:  intakeForm.CoordinatorLastName,
		HasClient:            intakeForm.HasClient,
		Goals: util.Map(intakeGoals, func(g db.ClientGoal) GoalItem {
			return GoalItem{
				ID:          &g.ID,
				Title:       g.Title,
				Description: g.Description,
			}
		}),
	}, nil
}

func (s *intakeService) UpdateIntakeForm(
	ctx context.Context,
	id string,
	req *UpdateIntakeFormRequest,
) (*UpdateIntakeFormResponse, error) {
	// Check if a client exists for this intake form
	intakeFormDetails, err := s.db.GetIntakeFormWithDetails(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "UpdateIntakeForm", "Failed to get intake form details", zap.Error(err))
		return nil, ErrInternal
	}

	// Build the update params
	params := db.UpdateIntakeFormParams{
		ID:                      id,
		FamilySituation:         req.FamilySituation,
		MainProvider:            req.MainProvider,
		Limitations:             req.Limitations,
		FocusAreas:              req.FocusAreas,
		Notes:                   req.Notes,
		EvaluationIntervalWeeks: util.IntToPointerInt32(req.EvaluationInterval),
		LocationID:              req.LocationID,
		CoordinatorID:           req.CoordinatorID,
	}

	// Handle date/time fields
	if req.IntakeDate != nil {
		params.IntakeDate = util.StrToPgtypeDate(*req.IntakeDate)
	}
	if req.IntakeTime != nil {
		params.IntakeTime = util.StrToPgtypeTime(*req.IntakeTime)
	}

	// Handle status enum field
	if req.Status != nil {
		params.Status = db.NullIntakeStatusEnum{
			IntakeStatusEnum: db.IntakeStatusEnum(*req.Status),
			Valid:            true,
		}
	}

	// Use transaction to update both intake form and client (if exists)
	err = s.db.UpdateIntakeFormTx(ctx, db.UpdateIntakeFormTxParams{
		IntakeForm:   params,
		UpdateClient: intakeFormDetails.HasClient,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateIntakeForm", "Failed to update intake form", zap.Error(err))
		return nil, ErrInternal
	}

	// Update goals (simple replacement logic: delete and recreate for now, or sophisticated diff)
	// For simplicity in early dev:
	if req.Goals != nil {
		err = s.db.ExecTx(ctx, func(q *db.Queries) error {
			// Delete existing goals
			existingGoals, err := q.ListGoalsByIntakeID(ctx, id)
			if err != nil {
				return err
			}
			for _, g := range existingGoals {
				if err := q.DeleteGoal(ctx, g.ID); err != nil {
					return err
				}
			}
			// Create new goals
			for _, g := range req.Goals {
				if err := q.CreateClientGoal(ctx, db.CreateClientGoalParams{
					ID:           nanoid.Generate(),
					IntakeFormID: id,
					Title:        g.Title,
					Description:  g.Description,
					ClientID:     nil, // Client will be linked later via LinkGoalsToClient if exists
				}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Error(ctx, "UpdateIntakeForm", "Failed to update goals", zap.Error(err))
			return nil, ErrInternal
		}
	}

	return &UpdateIntakeFormResponse{
		ID: id,
	}, nil
}

func (s *intakeService) GetIntakeStats(
	ctx context.Context,
) (*GetIntakeStatsResponse, error) {
	stats, err := s.db.GetIntakeStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetIntakeStats", "Failed to get intake statistics", zap.Error(err))
		return nil, ErrInternal
	}

	// Type assert numeric type from interface{}
	conversionPct := float64(stats.ConversionPercentage)

	return &GetIntakeStatsResponse{
		TotalCount:           int(stats.TotalCount),
		PendingCount:         int(stats.PendingCount),
		ConversionPercentage: conversionPct,
	}, nil
}
