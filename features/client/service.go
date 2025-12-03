package client

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"context"

	"go.uber.org/zap"
)

type clientService struct {
	db     *db.Store
	logger *logger.Logger
}

func NewClientService(db *db.Store, logger *logger.Logger) ClientService {
	return &clientService{db: db, logger: logger}
}

func (s *clientService) MoveClientToWaitingList(ctx context.Context, req *MoveClientToWaitingListRequest) (*MoveClientToWaitingListResponse, error) {
	if req.IntakeFormID == "" {
		s.logger.Error(ctx, "MoveClientToWaitingList", "Invalid request: empty intake form ID")
		return nil, ErrInvalidRequest
	}

	intakeForm, err := s.db.GetIntakeForm(ctx, req.IntakeFormID)
	if err != nil {
		s.logger.Error(ctx, "MoveClientToWaitingList", "Failed to get intake form", zap.Error(err))
		return nil, ErrIntakeFormNotFound
	}

	registrationForm, err := s.db.GetRegistrationForm(ctx, intakeForm.RegistrationFormID)
	if err != nil {
		s.logger.Error(ctx, "MoveClientToWaitingList", "Failed to get registration form", zap.Error(err))
		return nil, ErrRegistrationFormNotFound
	}

	// Generate unique client ID
	clientID := nanoid.Generate()

	// Prepare client creation parameters
	createClientParams := db.CreateClientParams{
		ID:                  clientID,
		FirstName:           registrationForm.FirstName,
		LastName:            registrationForm.LastName,
		Bsn:                 registrationForm.Bsn,
		DateOfBirth:         registrationForm.DateOfBirth,
		PhoneNumber:         nil, // Not available in intake/registration forms
		WaitingListPriority: db.WaitingListPriorityEnum(req.WaitingListPriority),
		Gender:              registrationForm.Gender,
		RegistrationFormID:  registrationForm.ID,
		IntakeFormID:        intakeForm.ID,
		CareType:            registrationForm.CareType,
		ReferringOrgID:      registrationForm.RefferingOrgID,
		Status:              db.ClientStatusEnumWaitingList,
		AssignedLocationID:  &intakeForm.LocationID,
		CoordinatorID:       &intakeForm.CoordinatorID,
		FamilySituation:     intakeForm.FamilySituation,
		Limitations:         intakeForm.Limitations,
		FocusAreas:          intakeForm.FocusAreas,
		Goals:               intakeForm.Goals,
		Notes:               intakeForm.Notes,
	}

	// Create the client
	createdClient, err := s.db.CreateClient(ctx, createClientParams)
	if err != nil {
		s.logger.Error(ctx, "MoveClientToWaitingList", "Failed to create client", zap.Error(err))
		return nil, ErrFailedToCreateClient
	}

	s.logger.Info(ctx, "MoveClientToWaitingList", "Client created successfully", zap.String("clientId", createdClient.ID))

	return &MoveClientToWaitingListResponse{
		ClientID: createdClient.ID,
	}, nil
}

func (s *clientService) MoveClientInCare(ctx context.Context) error {}
