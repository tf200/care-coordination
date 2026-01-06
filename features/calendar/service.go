package calendar

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/util"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type calendarService struct {
	store  db.StoreInterface
	logger logger.Logger
}

func NewCalendarService(store db.StoreInterface, logger logger.Logger) CalendarService {
	return &calendarService{
		store:  store,
		logger: logger,
	}
}

// Appointment methods

func (s *calendarService) CreateAppointment(ctx context.Context, organizerID string, req CreateAppointmentRequest) (*AppointmentResponse, error) {
	id := nanoid.Generate()

	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		params := db.CreateAppointmentParams{
			ID:             id,
			Title:          req.Title,
			Description:    &req.Description,
			StartTime:      pgtype.Timestamptz{Time: req.StartTime, Valid: true},
			EndTime:        pgtype.Timestamptz{Time: req.EndTime, Valid: true},
			Location:       &req.Location,
			OrganizerID:    organizerID,
			Status:         db.NullAppointmentStatusEnum{AppointmentStatusEnum: db.AppointmentStatusEnum(req.Status), Valid: req.Status != ""},
			Type:           db.AppointmentTypeEnum(req.Type),
			RecurrenceRule: &req.RecurrenceRule,
		}

		_, err := q.CreateAppointment(ctx, params)
		if err != nil {
			return err
		}

		for _, p := range req.Participants {
			err := q.AddAppointmentParticipant(ctx, db.AddAppointmentParticipantParams{
				AppointmentID:   id,
				ParticipantID:   p.ID,
				ParticipantType: db.ParticipantTypeEnum(p.Type),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "CreateAppointment", "Failed to create appointment", zap.Error(err))
		return nil, ErrInternal
	}

	return s.GetAppointment(ctx, id)
}

func (s *calendarService) GetAppointment(ctx context.Context, id string) (*AppointmentResponse, error) {
	var response *AppointmentResponse
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		appointment, err := q.GetAppointment(ctx, id)
		if err != nil {
			return err
		}

		participants, err := q.ListAppointmentParticipants(ctx, id)
		if err != nil {
			return err
		}

		participantDTOs := make([]ParticipantDTO, len(participants))
		for i, p := range participants {
			participantDTOs[i] = ParticipantDTO{
				ID:   p.ParticipantID,
				Type: ParticipantType(p.ParticipantType),
			}
		}

		response = &AppointmentResponse{
			ID:             appointment.ID,
			Title:          appointment.Title,
			Description:    util.HandleNilString(appointment.Description),
			StartTime:      appointment.StartTime.Time,
			EndTime:        appointment.EndTime.Time,
			Location:       util.HandleNilString(appointment.Location),
			OrganizerID:    appointment.OrganizerID,
			Status:         AppointmentStatus(appointment.Status.AppointmentStatusEnum),
			Type:           AppointmentType(appointment.Type),
			RecurrenceRule: util.HandleNilString(appointment.RecurrenceRule),
			Participants:   participantDTOs,
			CreatedAt:      appointment.CreatedAt.Time,
			UpdatedAt:      appointment.UpdatedAt.Time,
		}
		return nil
	})

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrAppointmentNotFound
		}
		s.logger.Error(ctx, "GetAppointment", "Failed to get appointment", zap.Error(err))
		return nil, ErrInternal
	}

	return response, nil
}

func (s *calendarService) UpdateAppointment(ctx context.Context, id string, req UpdateAppointmentRequest) (*AppointmentResponse, error) {
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		params := db.UpdateAppointmentParams{
			ID: id,
		}

		if req.Title != nil {
			params.Title = *req.Title
		}
		if req.Description != nil {
			params.Description = req.Description
		}
		if req.StartTime != nil {
			params.StartTime = pgtype.Timestamptz{Time: *req.StartTime, Valid: true}
		}
		if req.EndTime != nil {
			params.EndTime = pgtype.Timestamptz{Time: *req.EndTime, Valid: true}
		}
		if req.Location != nil {
			params.Location = req.Location
		}
		if req.Status != nil {
			params.Status = db.NullAppointmentStatusEnum{AppointmentStatusEnum: db.AppointmentStatusEnum(*req.Status), Valid: true}
		}
		if req.Type != nil {
			params.Type = db.NullAppointmentTypeEnum{AppointmentTypeEnum: db.AppointmentTypeEnum(*req.Type), Valid: true}
		}
		if req.RecurrenceRule != nil {
			params.RecurrenceRule = req.RecurrenceRule
		}

		_, err := q.UpdateAppointment(ctx, params)
		if err != nil {
			return err
		}

		if req.Participants != nil {
			err := q.RemoveAppointmentParticipants(ctx, id)
			if err != nil {
				return err
			}
			for _, p := range req.Participants {
				err := q.AddAppointmentParticipant(ctx, db.AddAppointmentParticipantParams{
					AppointmentID:   id,
					ParticipantID:   p.ID,
					ParticipantType: db.ParticipantTypeEnum(p.Type),
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "UpdateAppointment", "Failed to update appointment", zap.Error(err))
		return nil, ErrInternal
	}

	return s.GetAppointment(ctx, id)
}

func (s *calendarService) DeleteAppointment(ctx context.Context, id string) error {
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		return q.DeleteAppointment(ctx, id)
	})
	if err != nil {
		s.logger.Error(ctx, "DeleteAppointment", "Failed to delete appointment", zap.Error(err))
		return ErrInternal
	}
	return nil
}

func (s *calendarService) ListAppointments(ctx context.Context, userID string) ([]AppointmentResponse, error) {
	var responses []AppointmentResponse
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		appointments, err := q.ListAppointmentsByOrganizer(ctx, userID)
		if err != nil {
			return err
		}

		responses = make([]AppointmentResponse, 0, len(appointments))
		for _, a := range appointments {
			// Note: We avoid calling s.GetAppointment here to stay within the same transaction/queries object
			participants, err := q.ListAppointmentParticipants(ctx, a.ID)
			if err != nil {
				return err
			}

			participantDTOs := make([]ParticipantDTO, len(participants))
			for i, p := range participants {
				participantDTOs[i] = ParticipantDTO{
					ID:   p.ParticipantID,
					Type: ParticipantType(p.ParticipantType),
				}
			}

			responses = append(responses, AppointmentResponse{
				ID:             a.ID,
				Title:          a.Title,
				Description:    util.HandleNilString(a.Description),
				StartTime:      a.StartTime.Time,
				EndTime:        a.EndTime.Time,
				Location:       util.HandleNilString(a.Location),
				OrganizerID:    a.OrganizerID,
				Status:         AppointmentStatus(a.Status.AppointmentStatusEnum),
				Type:           AppointmentType(a.Type),
				RecurrenceRule: util.HandleNilString(a.RecurrenceRule),
				Participants:   participantDTOs,
				CreatedAt:      a.CreatedAt.Time,
				UpdatedAt:      a.UpdatedAt.Time,
			})
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "ListAppointments", "Failed to list appointments", zap.Error(err))
		return nil, ErrInternal
	}

	return responses, nil
}

// Reminder methods

func (s *calendarService) CreateReminder(ctx context.Context, userID string, req CreateReminderRequest) (*ReminderResponse, error) {
	var response *ReminderResponse
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		isCompleted := false
		params := db.CreateReminderParams{
			ID:          nanoid.Generate(),
			UserID:      userID,
			Title:       req.Title,
			Description: &req.Description,
			DueTime:     pgtype.Timestamptz{Time: req.DueTime, Valid: true},
			IsCompleted: &isCompleted,
		}

		reminder, err := q.CreateReminder(ctx, params)
		if err != nil {
			return err
		}

		response = &ReminderResponse{
			ID:          reminder.ID,
			UserID:      reminder.UserID,
			Title:       reminder.Title,
			Description: util.HandleNilString(reminder.Description),
			DueTime:     reminder.DueTime.Time,
			IsCompleted: util.HandleNilBool(reminder.IsCompleted),
			CreatedAt:   reminder.CreatedAt.Time,
			UpdatedAt:   reminder.UpdatedAt.Time,
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "CreateReminder", "Failed to create reminder", zap.Error(err))
		return nil, ErrInternal
	}

	return response, nil
}

func (s *calendarService) GetReminder(ctx context.Context, id string) (*ReminderResponse, error) {
	var response *ReminderResponse
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		reminder, err := q.GetReminder(ctx, id)
		if err != nil {
			return err
		}

		response = &ReminderResponse{
			ID:          reminder.ID,
			UserID:      reminder.UserID,
			Title:       reminder.Title,
			Description: util.HandleNilString(reminder.Description),
			DueTime:     reminder.DueTime.Time,
			IsCompleted: util.HandleNilBool(reminder.IsCompleted),
			CreatedAt:   reminder.CreatedAt.Time,
			UpdatedAt:   reminder.UpdatedAt.Time,
		}
		return nil
	})

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrReminderNotFound
		}
		s.logger.Error(ctx, "GetReminder", "Failed to get reminder", zap.Error(err))
		return nil, ErrInternal
	}

	return response, nil
}

func (s *calendarService) UpdateReminder(ctx context.Context, id string, completed bool) (*ReminderResponse, error) {
	var response *ReminderResponse
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		params := db.UpdateReminderParams{
			ID:          id,
			IsCompleted: &completed,
		}

		reminder, err := q.UpdateReminder(ctx, params)
		if err != nil {
			return err
		}

		response = &ReminderResponse{
			ID:          reminder.ID,
			UserID:      reminder.UserID,
			Title:       reminder.Title,
			Description: util.HandleNilString(reminder.Description),
			DueTime:     reminder.DueTime.Time,
			IsCompleted: util.HandleNilBool(reminder.IsCompleted),
			CreatedAt:   reminder.CreatedAt.Time,
			UpdatedAt:   reminder.UpdatedAt.Time,
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "UpdateReminder", "Failed to update reminder", zap.Error(err))
		return nil, ErrInternal
	}

	return response, nil
}

func (s *calendarService) DeleteReminder(ctx context.Context, id string) error {
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		return q.DeleteReminder(ctx, id)
	})
	if err != nil {
		s.logger.Error(ctx, "DeleteReminder", "Failed to delete reminder", zap.Error(err))
		return ErrInternal
	}
	return nil
}

func (s *calendarService) ListReminders(ctx context.Context, userID string) ([]ReminderResponse, error) {
	var responses []ReminderResponse
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		reminders, err := q.ListRemindersByUser(ctx, userID)
		if err != nil {
			return err
		}

		responses = make([]ReminderResponse, len(reminders))
		for i, r := range reminders {
			responses[i] = ReminderResponse{
				ID:          r.ID,
				UserID:      r.UserID,
				Title:       r.Title,
				Description: util.HandleNilString(r.Description),
				DueTime:     r.DueTime.Time,
				IsCompleted: util.HandleNilBool(r.IsCompleted),
				CreatedAt:   r.CreatedAt.Time,
				UpdatedAt:   r.UpdatedAt.Time,
			}
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "ListReminders", "Failed to list reminders", zap.Error(err))
		return nil, ErrInternal
	}

	return responses, nil
}

func (s *calendarService) GetCalendarView(ctx context.Context, userID string, startTime, endTime time.Time) ([]CalendarEventDTO, error) {
	var events []CalendarEventDTO
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		appointments, err := q.ListAppointmentsByRange(ctx, db.ListAppointmentsByRangeParams{
			OrganizerID: userID,
			StartTime:   pgtype.Timestamptz{Time: startTime, Valid: true},
			EndTime:     pgtype.Timestamptz{Time: endTime, Valid: true},
		})
		if err != nil {
			return err
		}

		reminders, err := q.ListRemindersByRange(ctx, db.ListRemindersByRangeParams{
			UserID:    userID,
			StartTime: pgtype.Timestamptz{Time: startTime, Valid: true},
			EndTime:   pgtype.Timestamptz{Time: endTime, Valid: true},
		})
		if err != nil {
			return err
		}

		events = make([]CalendarEventDTO, 0, len(appointments)+len(reminders))

		for _, a := range appointments {
			endTimeValue := a.EndTime.Time
			events = append(events, CalendarEventDTO{
				ID:              a.ID,
				Title:           a.Title,
				Start:           a.StartTime.Time,
				End:             &endTimeValue,
				AllDay:          false,
				Type:            "appointment",
				BackgroundColor: "#3788d8", // Blue
				ExtendedProps: CalendarExtendedProps{
					Description: util.HandleNilString(a.Description),
					Location:    util.HandleNilString(a.Location),
					Status:      string(a.Status.AppointmentStatusEnum),
					Type:        string(a.Type),
				},
			})
		}

		for _, r := range reminders {
			events = append(events, CalendarEventDTO{
				ID:              r.ID,
				Title:           r.Title,
				Start:           r.DueTime.Time,
				AllDay:          false,
				Type:            "reminder",
				BackgroundColor: "#28a745", // Green
				ExtendedProps: CalendarExtendedProps{
					Description: util.HandleNilString(r.Description),
					IsCompleted: util.HandleNilBool(r.IsCompleted),
				},
			})
		}
		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "GetCalendarView", "Failed to list items for view", zap.Error(err))
		return nil, ErrInternal
	}

	return events, nil
}
