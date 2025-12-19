package registration

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

func (s *registrationService) CreateRegistrationForm(
	ctx context.Context,
	req *CreateRegistrationFormRequest,
) (*CreateRegistrationFormResponse, error) {
	id := nanoid.Generate()
	err := s.db.CreateRegistrationForm(ctx, db.CreateRegistrationFormParams{
		ID:                 id,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Bsn:                req.BSN,
		DateOfBirth:        util.StrToPgtypeDate(req.DateOfBirth),
		PhoneNumber:        req.PhoneNumber,
		RefferingOrgID:     req.RefferingOrgID,
		Gender:             db.GenderEnum(req.Gender),
		CareType:           db.CareTypeEnum(req.CareType),
		RegistrationDate:   util.StrToPgtypeDate(req.RegistrationDate),
		RegistrationReason: req.RegistrationReason,
		AdditionalNotes:    req.AdditionalNotes,
		AttachmentIds:      req.AttachmentIDs,
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"CreateRegistrationForm",
			"Failed to create registration form",
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	return &CreateRegistrationFormResponse{
		ID: id,
	}, nil
}

func (s *registrationService) ListRegistrationForms(
	ctx context.Context,
	req *ListRegistrationFormsRequest,
) (*resp.PaginationResponse[ListRegistrationFormsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	var registrationForms []db.ListRegistrationFormsRow
	err := s.db.ExecTx(ctx, func(q *db.Queries) error {
		var err error
		registrationForms, err = q.ListRegistrationForms(ctx, db.ListRegistrationFormsParams{
			Limit:           limit,
			Offset:          offset,
			Search:          req.Search,
			Status:          req.Status,
			IntakeCompleted: req.IntakeCompleted,
		})
		return err
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"ListRegistrationForms",
			"Failed to list registration forms",
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	listRegistrationFormsResponse := []ListRegistrationFormsResponse{}
	for _, registrationForm := range registrationForms {
		status := ""
		if registrationForm.Status.Valid {
			status = string(registrationForm.Status.RegistrationStatusEnum)
		}

		listRegistrationFormsResponse = append(
			listRegistrationFormsResponse,
			ListRegistrationFormsResponse{
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
				IntakeCompleted:     registrationForm.IntakeCompleted,
			},
		)
	}
	totalCount := 0
	if len(registrationForms) > 0 {
		totalCount = int(registrationForms[0].TotalCount)
	}
	// Use page and pageSize (not offset and limit) for correct pagination metadata
	result := resp.PagRespWithParams(listRegistrationFormsResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *registrationService) UpdateRegistrationForm(
	ctx context.Context,
	id string,
	req *UpdateRegistrationFormRequest,
) (*UpdateRegistrationFormResponse, error) {
	// Check if a client exists for this registration form
	regFormDetails, err := s.db.GetRegistrationFormWithDetails(ctx, id)
	if err != nil {
		s.logger.Error(
			ctx,
			"UpdateRegistrationForm",
			"Failed to get registration form details",
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	// Build the update params - only set fields that are provided
	params := db.UpdateRegistrationFormParams{
		ID:                 id,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Bsn:                req.BSN,
		PhoneNumber:        req.PhoneNumber,
		RefferingOrgID:     req.RefferingOrgID,
		RegistrationReason: req.RegistrationReason,
		AdditionalNotes:    req.AdditionalNotes,
		AttachmentIds:      req.AttachmentIDs,
	}

	// Handle date fields
	if req.DateOfBirth != nil {
		params.DateOfBirth = util.StrToPgtypeDate(*req.DateOfBirth)
	}
	if req.RegistrationDate != nil {
		params.RegistrationDate = util.StrToPgtypeDate(*req.RegistrationDate)
	}

	// Handle enum fields
	if req.Gender != nil {
		params.Gender = db.NullGenderEnum{
			GenderEnum: db.GenderEnum(*req.Gender),
			Valid:      true,
		}
	}
	if req.CareType != nil {
		params.CareType = db.NullCareTypeEnum{
			CareTypeEnum: db.CareTypeEnum(*req.CareType),
			Valid:        true,
		}
	}
	if req.Status != nil {
		params.Status = db.NullRegistrationStatusEnum{
			RegistrationStatusEnum: db.RegistrationStatusEnum(*req.Status),
			Valid:                  true,
		}
	}

	// Use transaction to update both registration form and client (if exists)
	err = s.db.UpdateRegistrationFormTx(ctx, db.UpdateRegistrationFormTxParams{
		RegistrationForm: params,
		UpdateClient:     regFormDetails.HasClient,
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"UpdateRegistrationForm",
			"Failed to update registration form",
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	return &UpdateRegistrationFormResponse{
		ID: id,
	}, nil
}

func (s *registrationService) GetRegistrationForm(
	ctx context.Context,
	id string,
) (*GetRegistrationFormResponse, error) {
	regForm, err := s.db.GetRegistrationFormWithDetails(ctx, id)
	if err != nil {
		s.logger.Error(
			ctx,
			"GetRegistrationForm",
			"Failed to get registration form",
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	status := ""
	if regForm.Status.Valid {
		status = string(regForm.Status.RegistrationStatusEnum)
	}

	return &GetRegistrationFormResponse{
		ID:                 regForm.ID,
		FirstName:          regForm.FirstName,
		LastName:           regForm.LastName,
		Bsn:                regForm.Bsn,
		Gender:             string(regForm.Gender),
		DateOfBirth:        regForm.DateOfBirth.Time,
		RefferingOrgID:     regForm.RefferingOrgID,
		OrgName:            regForm.OrgName,
		OrgContactPerson:   regForm.OrgContactPerson,
		OrgPhoneNumber:     regForm.OrgPhoneNumber,
		OrgEmail:           regForm.OrgEmail,
		CareType:           string(regForm.CareType),
		RegistrationDate:   regForm.RegistrationDate.Time,
		RegistrationReason: regForm.RegistrationReason,
		AdditionalNotes:    regForm.AdditionalNotes,
		AttachmentIDs:      regForm.AttachmentIds,
		Status:             &status,
		IntakeCompleted:    regForm.IntakeCompleted,
		HasClient:          regForm.HasClient,
	}, nil
}

func (s *registrationService) DeleteRegistrationForm(
	ctx context.Context,
	id string,
) (*DeleteRegistrationFormResponse, error) {
	err := s.db.SoftDeleteRegistrationForm(ctx, id)
	if err != nil {
		s.logger.Error(
			ctx,
			"DeleteRegistrationForm",
			"Failed to delete registration form",
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	return &DeleteRegistrationFormResponse{
		ID: id,
	}, nil
}
