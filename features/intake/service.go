package intake

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

type intakeService struct {
	db     *db.Store
	logger *logger.Logger
}

func NewIntakeService(db *db.Store, logger *logger.Logger) IntakeService {
	return &intakeService{
		db:     db,
		logger: logger,
	}
}

func (s *intakeService) CreateIntakeForm(ctx context.Context, req *CreateIntakeFormRequest) (*CreateIntakeFormResponse, error) {
	id := nanoid.Generate()
	err := s.db.CreateIntakeForm(ctx, db.CreateIntakeFormParams{
		ID:                 id,
		RegistrationFormID: req.RegistrationFormID,
		IntakeDate:         util.StrToPgtypeDate(req.IntakeDate),
		IntakeTime:         util.StrToPgtypeTime(req.IntakeTime),
		LocationID:         req.LocationID,
		CoordinatorID:      req.CoordinatorID,
		FamilySituation:    req.FamilySituation,
		MainProvider:       req.MainProvider,
		Limitations:        req.Limitations,
		FocusAreas:         req.FocusAreas,
		Goals:              req.Goals,
		Notes:              req.Notes,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateIntakeForm", "Failed to create intake form", zap.Error(err))
		return nil, ErrInternal
	}
	return &CreateIntakeFormResponse{
		ID: id,
	}, nil
}

func (s *intakeService) ListIntakeForms(ctx context.Context, req *ListIntakeFormsRequest) (*resp.PaginationResponse[ListIntakeFormsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)
	intakeForms, err := s.db.ListIntakeForms(ctx, db.ListIntakeFormsParams{
		Limit:   limit,
		Offset:  offset,
		Column3: *req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListIntakeForms", "Failed to list intake forms", zap.Error(err))
		return nil, ErrInternal
	}
	listIntakeFormsResponse := []ListIntakeFormsResponse{}
	for _, intakeForm := range intakeForms {
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
