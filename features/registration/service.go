package registration

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
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
		OrgName:            req.OrgName,
		OrgContactPerson:   req.OrgContactPerson,
		OrgPhoneNumber:     req.OrgPhoneNumber,
		OrgEmail:           req.OrgEmail,
		CareType:           db.CareTypeEnum(req.CareType),
		CoordinatorID:      req.CoordinatorID,
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

func (s *registrationService) ListRegistrationForms(ctx context.Context) ([]ListRegistrationFormsResponse, error) {
	registrationForms, err := s.db.ListRegistrationForms(ctx)
	if err != nil {
		s.logger.Error(ctx, "ListRegistrationForms", "Failed to list registration forms", zap.Error(err))
		return nil, ErrInternal
	}
	listRegistrationFormsResponse := []ListRegistrationFormsResponse{}
	for _, registrationForm := range registrationForms {
		listRegistrationFormsResponse = append(listRegistrationFormsResponse, ListRegistrationFormsResponse{
			ID:                   registrationForm.ID,
			FirstName:            registrationForm.FirstName,
			LastName:             registrationForm.LastName,
			Bsn:                  registrationForm.Bsn,
			DateOfBirth:          registrationForm.DateOfBirth.Time,
			OrgName:              registrationForm.OrgName,
			CareType:             string(registrationForm.CareType),
			CoordinatorID:        registrationForm.CoordinatorID,
			CoordinatorFirstName: registrationForm.CoordinatorFirstName,
			CoordinatorLastName:  registrationForm.CoordinatorLastName,
			RegistrationDate:     registrationForm.RegistrationDate.Time,
			RegistrationReason:   registrationForm.RegistrationReason,
			NumberOfAttachments:  len(registrationForm.AttachmentIds),
		})
	}
	return listRegistrationFormsResponse, nil
}
