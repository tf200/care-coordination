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

func (s *clientService) MoveClientInCare(ctx context.Context, clientID string, req *MoveClientInCareRequest) (*MoveClientInCareResponse, error) {
	client, err := s.db.GetClientByID(ctx, clientID)
	if err != nil {
		s.logger.Error(ctx, "MoveClientInCare", "Failed to get client", zap.Error(err))
		return nil, ErrClientNotFound
	}

	// Validate client is on waiting list
	if client.Status != db.ClientStatusEnumWaitingList {
		s.logger.Error(ctx, "MoveClientInCare", "Client must be on waiting list to move to in care", zap.String("currentStatus", string(client.Status)))
		return nil, ErrInvalidClientStatus
	}

	// Validate ambulatory weekly hours based on care type
	isAmbulatory := client.CareType == db.CareTypeEnumAmbulatoryCare

	if isAmbulatory && (req.AmbulatoryWeeklyHours == nil || *req.AmbulatoryWeeklyHours <= 0) {
		s.logger.Error(ctx, "MoveClientInCare", "Ambulatory weekly hours required for ambulatory care")
		return nil, ErrAmbulatoryHoursRequired
	}

	if !isAmbulatory && req.AmbulatoryWeeklyHours != nil {
		s.logger.Error(ctx, "MoveClientInCare", "Ambulatory weekly hours should only be set for ambulatory care", zap.String("careType", string(client.CareType)))
		return nil, ErrAmbulatoryHoursNotAllowed
	}

	updateParams := db.UpdateClientParams{
		ID:                    client.ID,
		Status:                db.ClientStatusEnumInCare,
		AmbulatoryWeeklyHours: req.AmbulatoryWeeklyHours,
		CareStartDate:         req.CareStartDate,
		CareEndDate:           req.CareEndDate,
	}

	updatedClient, err := s.db.UpdateClient(ctx, updateParams)
	if err != nil {
		s.logger.Error(ctx, "MoveClientInCare", "Failed to update client status", zap.Error(err))
		return nil, ErrInternal
	}

	s.logger.Info(ctx, "MoveClientInCare", "Client moved to in care successfully", zap.String("clientId", updatedClient))

	return &MoveClientInCareResponse{
		ClientID: updatedClient,
	}, nil
}
