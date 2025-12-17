package client

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"context"
	"math/rand"
	"time"

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
		PhoneNumber:         registrationForm.PhoneNumber,
		WaitingListPriority: db.WaitingListPriorityEnum(req.WaitingListPriority),
		Gender:              registrationForm.Gender,
		RegistrationFormID:  registrationForm.ID,
		IntakeFormID:        intakeForm.ID,
		CareType:            registrationForm.CareType,
		ReferringOrgID:      registrationForm.RefferingOrgID,
		Status:              db.ClientStatusEnumWaitingList,
		AssignedLocationID:  intakeForm.LocationID,
		CoordinatorID:       intakeForm.CoordinatorID,
		FamilySituation:     intakeForm.FamilySituation,
		Limitations:         intakeForm.Limitations,
		FocusAreas:          intakeForm.FocusAreas,
		Goals:               intakeForm.Goals,
		Notes:               intakeForm.Notes,
	}

	// Create the client and update intake form status in a transaction
	result, err := s.db.MoveClientToWaitingListTx(ctx, db.MoveClientToWaitingListTxParams{
		Client:              createClientParams,
		IntakeFormID:        intakeForm.ID,
		IntakeFormNewStatus: db.IntakeStatusEnumCompleted,
	})
	if err != nil {
		s.logger.Error(ctx, "MoveClientToWaitingList", "Failed to create client and update intake form", zap.Error(err))
		return nil, ErrFailedToCreateClient
	}

	s.logger.Info(ctx, "MoveClientToWaitingList", "Client created and intake form completed successfully", zap.String("clientId", result.ClientID))

	return &MoveClientToWaitingListResponse{
		ClientID: result.ClientID,
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
		Status:                db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumInCare, Valid: true},
		AmbulatoryWeeklyHours: req.AmbulatoryWeeklyHours,
		CareStartDate:         util.StrToPgtypeDate(req.CareStartDate),
		CareEndDate:           util.StrToPgtypeDate(req.CareEndDate),
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

func (s *clientService) StartDischarge(ctx context.Context, clientID string, req *StartDischargeRequest) (*StartDischargeResponse, error) {
	client, err := s.db.GetClientByID(ctx, clientID)
	if err != nil {
		s.logger.Error(ctx, "StartDischarge", "Failed to get client", zap.Error(err))
		return nil, ErrClientNotFound
	}

	// Validate client is in care
	if client.Status != db.ClientStatusEnumInCare {
		s.logger.Error(ctx, "StartDischarge", "Client must be in care to start discharge", zap.String("currentStatus", string(client.Status)))
		return nil, ErrClientNotInCare
	}

	// Validate discharge has not already been started
	if client.DischargeStatus.Valid {
		s.logger.Error(ctx, "StartDischarge", "Discharge already started for this client", zap.String("dischargeStatus", string(client.DischargeStatus.DischargeStatusEnum)))
		return nil, ErrDischargeAlreadyStarted
	}

	updateParams := db.UpdateClientParams{
		ID:                 client.ID,
		Status:             db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumInCare, Valid: true}, // Client remains in care during phase 1
		DischargeDate:      util.StrToPgtypeDate(req.DischargeDate),
		ReasonForDischarge: db.NullDischargeReasonEnum{DischargeReasonEnum: db.DischargeReasonEnum(req.ReasonForDischarge), Valid: true},
		DischargeStatus:    db.NullDischargeStatusEnum{DischargeStatusEnum: db.DischargeStatusEnumInProgress, Valid: true},
	}

	updatedClient, err := s.db.UpdateClient(ctx, updateParams)
	if err != nil {
		s.logger.Error(ctx, "StartDischarge", "Failed to update client", zap.Error(err))
		return nil, ErrInternal
	}

	s.logger.Info(ctx, "StartDischarge", "Discharge process started for client", zap.String("clientId", updatedClient))

	return &StartDischargeResponse{
		ClientID: updatedClient,
	}, nil
}

func (s *clientService) CompleteDischarge(ctx context.Context, clientID string, req *CompleteDischargeRequest) (*CompleteDischargeResponse, error) {
	client, err := s.db.GetClientByID(ctx, clientID)
	if err != nil {
		s.logger.Error(ctx, "CompleteDischarge", "Failed to get client", zap.Error(err))
		return nil, ErrClientNotFound
	}

	// Validate client is in care
	if client.Status != db.ClientStatusEnumInCare {
		s.logger.Error(ctx, "CompleteDischarge", "Client must be in care to complete discharge", zap.String("currentStatus", string(client.Status)))
		return nil, ErrClientNotInCare
	}

	// Validate discharge has been started (phase 1 completed)
	if !client.DischargeStatus.Valid || client.DischargeStatus.DischargeStatusEnum != db.DischargeStatusEnumInProgress {
		s.logger.Error(ctx, "CompleteDischarge", "Discharge must be started before completing")
		return nil, ErrDischargeNotStarted
	}

	updateParams := db.UpdateClientParams{
		ID:                     client.ID,
		Status:                 db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumDischarged, Valid: true}, // Now move to discharged
		ClosingReport:          &req.ClosingReport,
		EvaluationReport:       &req.EvaluationReport,
		DischargeAttachmentIds: req.DischargeAttachmentIDs,
		DischargeStatus:        db.NullDischargeStatusEnum{DischargeStatusEnum: db.DischargeStatusEnumCompleted, Valid: true},
	}

	updatedClient, err := s.db.UpdateClient(ctx, updateParams)
	if err != nil {
		s.logger.Error(ctx, "CompleteDischarge", "Failed to update client", zap.Error(err))
		return nil, ErrInternal
	}

	s.logger.Info(ctx, "CompleteDischarge", "Client discharge completed", zap.String("clientId", updatedClient))

	return &CompleteDischargeResponse{
		ClientID: updatedClient,
	}, nil
}

func (s *clientService) ListWaitingListClients(ctx context.Context, req *ListWaitingListClientsRequest) (*resp.PaginationResponse[ListWaitingListClientsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	clients, err := s.db.ListWaitingListClients(ctx, db.ListWaitingListClientsParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListWaitingListClients", "Failed to list waiting list clients", zap.Error(err))
		return nil, ErrInternal
	}

	listClientsResponse := []ListWaitingListClientsResponse{}
	totalCount := 0

	for _, client := range clients {
		listClientsResponse = append(listClientsResponse, ListWaitingListClientsResponse{
			ID:                   client.ID,
			FirstName:            client.FirstName,
			LastName:             client.LastName,
			Bsn:                  client.Bsn,
			DateOfBirth:          util.PgtypeDateToStr(client.DateOfBirth),
			PhoneNumber:          client.PhoneNumber,
			Gender:               string(client.Gender),
			CareType:             string(client.CareType),
			WaitingListPriority:  string(client.WaitingListPriority),
			FocusAreas:           client.FocusAreas,
			Notes:                client.Notes,
			CreatedAt:            util.PgtypeTimestampToStr(client.CreatedAt),
			LocationID:           client.LocationID,
			LocationName:         client.LocationName,
			CoordinatorID:        client.CoordinatorID,
			CoordinatorFirstName: client.CoordinatorFirstName,
			CoordinatorLastName:  client.CoordinatorLastName,
			ReferringOrgName:     client.ReferringOrgName,
		})
		if totalCount == 0 {
			totalCount = int(client.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listClientsResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *clientService) ListInCareClients(ctx context.Context, req *ListInCareClientsRequest) (*resp.PaginationResponse[ListInCareClientsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	clients, err := s.db.ListInCareClients(ctx, db.ListInCareClientsParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListInCareClients", "Failed to list in care clients", zap.Error(err))
		return nil, ErrInternal
	}

	listClientsResponse := []ListInCareClientsResponse{}
	totalCount := 0
	now := time.Now()

	for _, client := range clients {
		response := ListInCareClientsResponse{
			ID:                   client.ID,
			FirstName:            client.FirstName,
			LastName:             client.LastName,
			Bsn:                  client.Bsn,
			DateOfBirth:          util.PgtypeDateToStr(client.DateOfBirth),
			PhoneNumber:          client.PhoneNumber,
			Gender:               string(client.Gender),
			CareType:             string(client.CareType),
			CareStartDate:        util.PgtypeDateToStr(client.CareStartDate),
			CareEndDate:          util.PgtypeDateToStr(client.CareEndDate),
			LocationID:           client.LocationID,
			LocationName:         client.LocationName,
			CoordinatorID:        client.CoordinatorID,
			CoordinatorFirstName: client.CoordinatorFirstName,
			CoordinatorLastName:  client.CoordinatorLastName,
			ReferringOrgName:     client.ReferringOrgName,
		}

		// Calculate weeks in accommodation or used ambulatory hours based on care type
		if client.CareType == db.CareTypeEnumAmbulatoryCare {
			// For ambulatory care: generate random used hours for demo purposes
			randomHours := rand.Intn(100) + 1 // 1-100 hours
			response.UsedAmbulatoryHours = &randomHours
		} else {
			// For protected_living, semi_independent_living, independent_assisted_living
			// Calculate weeks from care start date
			if client.CareStartDate.Valid {
				startDate := client.CareStartDate.Time
				duration := now.Sub(startDate)
				weeks := int(duration.Hours() / 24 / 7)
				response.WeeksInAccommodation = &weeks
			}
		}

		listClientsResponse = append(listClientsResponse, response)
		if totalCount == 0 {
			totalCount = int(client.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listClientsResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *clientService) ListDischargedClients(ctx context.Context, req *ListDischargedClientsRequest) (*resp.PaginationResponse[ListDischargedClientsResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	// Build discharge status filter
	var dischargeStatusFilter db.NullDischargeStatusEnum
	if req.DischargeStatus != nil {
		dischargeStatusFilter = db.NullDischargeStatusEnum{
			DischargeStatusEnum: db.DischargeStatusEnum(*req.DischargeStatus),
			Valid:               true,
		}
	}

	clients, err := s.db.ListDischargedClients(ctx, db.ListDischargedClientsParams{
		Limit:           limit,
		Offset:          offset,
		Search:          req.Search,
		DischargeStatus: dischargeStatusFilter,
	})
	if err != nil {
		s.logger.Error(ctx, "ListDischargedClients", "Failed to list discharged clients", zap.Error(err))
		return nil, ErrInternal
	}

	listClientsResponse := []ListDischargedClientsResponse{}
	totalCount := 0

	for _, client := range clients {
		response := ListDischargedClientsResponse{
			ID:                   client.ID,
			FirstName:            client.FirstName,
			LastName:             client.LastName,
			Bsn:                  client.Bsn,
			DateOfBirth:          util.PgtypeDateToStr(client.DateOfBirth),
			PhoneNumber:          client.PhoneNumber,
			Gender:               string(client.Gender),
			CareType:             string(client.CareType),
			CareStartDate:        util.PgtypeDateToStr(client.CareStartDate),
			DischargeDate:        util.PgtypeDateToStr(client.DischargeDate),
			ReasonForDischarge:   string(client.ReasonForDischarge.DischargeReasonEnum),
			DischargeStatus:      string(client.DischargeStatus.DischargeStatusEnum),
			ClosingReport:        client.ClosingReport,
			EvaluationReport:     client.EvaluationReport,
			LocationID:           client.LocationID,
			LocationName:         client.LocationName,
			CoordinatorID:        client.CoordinatorID,
			CoordinatorFirstName: client.CoordinatorFirstName,
			CoordinatorLastName:  client.CoordinatorLastName,
			ReferringOrgName:     client.ReferringOrgName,
		}

		listClientsResponse = append(listClientsResponse, response)
		if totalCount == 0 {
			totalCount = int(client.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listClientsResponse, totalCount, page, pageSize)
	return &result, nil
}
