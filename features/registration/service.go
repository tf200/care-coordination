package registration

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type registrationService struct {
	db     *db.Store
	logger *logger.Logger
}

func NewRegistrationService(db *db.Store, logger *logger.Logger) RegistrationService {
	return &registrationService{
		db:     db,
		logger: logger,
	}
}

func (s *registrationService) CreateRegistrationForm(ctx context.Context, req *CreateRegistrationFormRequest) (*CreateRegistrationFormResponse, error) {
	id := nanoid.Generate()
	err := s.db.CreateRegistrationForm(ctx, db.CreateRegistrationFormParams{
		ID:                 id,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Bsn:                req.BSN,
		DateOfBirth:        pgtype.Date{Time: req.DateOfBirth, Valid: true},
		RefferingOrgID:     req.RefferingOrgID,
		Gender:             db.GenderEnum(req.Gender),
		CareType:           db.CareTypeEnum(req.CareType),
		RegistrationDate:   pgtype.Timestamp{Time: req.RegistrationDate, Valid: true},
		RegistrationReason: req.RegistrationReason,
		AdditionalNotes:    req.AdditionalNotes,
		AttachmentIds:      req.AttachmentIDs,
	})
	if err != nil {
		s.logger.Error(ctx, "CreateRegistrationForm", "Failed to create registration form", zap.Error(err))
		return nil, ErrInternal
	}
	return &CreateRegistrationFormResponse{
		ID: id,
	}, nil
}

func (s *registrationService) ListRegistrationForms(ctx context.Context, req *ListRegistrationFormsRequest) (*resp.PaginationResponse[ListRegistrationFormsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)
	registrationForms, err := s.db.ListRegistrationForms(ctx, db.ListRegistrationFormsParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListRegistrationForms", "Failed to list registration forms", zap.Error(err))
		return nil, ErrInternal
	}
	listRegistrationFormsResponse := []ListRegistrationFormsResponse{}
	for _, registrationForm := range registrationForms {
		status := ""
		if registrationForm.Status.Valid {
			status = string(registrationForm.Status.RegistrationStatusEnum)
		}

		listRegistrationFormsResponse = append(listRegistrationFormsResponse, ListRegistrationFormsResponse{
			ID:                  registrationForm.ID,
			FirstName:           registrationForm.FirstName,
			LastName:            registrationForm.LastName,
			Bsn:                 registrationForm.Bsn,
			DateOfBirth:         registrationForm.DateOfBirth.Time,
			RefferingOrgID:      registrationForm.RefferingOrgID,
			OrgName:             registrationForm.OrgName,
			OrgContactPerson:    registrationForm.OrgContactPerson,
			OrgPhoneNumber:      registrationForm.OrgPhoneNumber,
			OrgEmail:            registrationForm.OrgEmail,
			CareType:            string(registrationForm.CareType),
			RegistrationDate:    registrationForm.RegistrationDate.Time,
			RegistrationReason:  registrationForm.RegistrationReason,
			AdditionalNotes:     registrationForm.AdditionalNotes,
			NumberOfAttachments: len(registrationForm.AttachmentIds),
			Status:              &status,
		})
	}
	totalCount := 0
	if len(registrationForms) > 0 {
		totalCount = int(registrationForms[0].TotalCount)
	}
	// Use page and pageSize (not offset and limit) for correct pagination metadata
	result := resp.PagRespWithParams(listRegistrationFormsResponse, totalCount, page, pageSize)
	return &result, nil
}
