package main

import (
	"care-cordination/features/notification"
	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/util"
	"care-cordination/lib/websocket"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	// How often to run the scheduler
	tickInterval = 5 * time.Minute

	// Prevent duplicate notifications by tracking sent ones
	notificationCooldown = 30 * time.Minute
)

// sentNotifications tracks recently sent notifications to avoid duplicates
var sentNotifications = make(map[string]time.Time)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("cannot load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	l := logger.NewLogger(cfg.Environment)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	l.Info(ctx, "worker", "Starting notification background worker")

	// 3. Initialize Database Connection
	poolConfig, err := pgxpool.ParseConfig(cfg.DBSource)
	if err != nil {
		l.Error(ctx, "worker", "cannot parse db config", zap.Error(err))
		os.Exit(1)
	}

	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec

	connPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		l.Error(ctx, "worker", "cannot connect to db", zap.Error(err))
		os.Exit(1)
	}
	defer connPool.Close()

	// 4. Initialize Dependencies
	store := db.NewStore(connPool)

	// Initialize WebSocket Hub (for real-time delivery if worker runs with API)
	wsHub := websocket.NewHub(l)
	go wsHub.Run()

	notificationService := notification.NewNotificationService(store, wsHub, l)

	// 5. Create the worker
	worker := &NotificationWorker{
		store:               store,
		notificationService: notificationService,
		logger:              l,
	}

	// 6. Run the ticker
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	l.Info(ctx, "worker", "Worker started, running every", zap.Duration("interval", tickInterval))

	// Run immediately on start
	worker.Run(ctx)

	for {
		select {
		case <-ticker.C:
			worker.Run(ctx)
		case <-ctx.Done():
			l.Info(ctx, "worker", "Shutdown signal received, stopping worker")
			return
		}
	}
}

// NotificationWorker handles scheduled notification checks
type NotificationWorker struct {
	store               *db.Store
	notificationService notification.NotificationService
	logger              logger.Logger
}

// Run executes all notification checks
func (w *NotificationWorker) Run(ctx context.Context) {
	w.logger.Info(ctx, "worker", "Running scheduled notification checks")

	// Clean up old sent notification records
	w.cleanupSentNotifications()

	// Check for various scheduled notifications
	w.checkUpcomingAppointments(ctx)
	w.checkEvaluationsDueSoon(ctx)
	w.checkPendingReminders(ctx)

	w.logger.Info(ctx, "worker", "Scheduled notification checks completed")
}

// cleanupSentNotifications removes old entries from the sent tracking map
func (w *NotificationWorker) cleanupSentNotifications() {
	now := time.Now()
	for key, sentAt := range sentNotifications {
		if now.Sub(sentAt) > notificationCooldown {
			delete(sentNotifications, key)
		}
	}
}

// shouldSendNotification checks if we should send a notification (not recently sent)
func shouldSendNotification(key string) bool {
	if sentAt, exists := sentNotifications[key]; exists {
		if time.Since(sentAt) < notificationCooldown {
			return false
		}
	}
	sentNotifications[key] = time.Now()
	return true
}

// checkUpcomingAppointments sends reminders for appointments starting soon
func (w *NotificationWorker) checkUpcomingAppointments(ctx context.Context) {
	appointments, err := w.store.GetUpcomingAppointments(ctx)
	if err != nil {
		w.logger.Error(ctx, "worker", "Failed to get upcoming appointments", zap.Error(err))
		return
	}

	for _, apt := range appointments {
		key := fmt.Sprintf("appointment:%s", apt.ID)
		if !shouldSendNotification(key) {
			continue
		}

		resourceType := notification.ResourceTypeAppointment
		resourceID := apt.ID

		// Calculate time until appointment
		timeUntil := time.Until(apt.StartTime.Time)
		minutesUntil := int(timeUntil.Minutes())

		w.notificationService.Enqueue(&notification.CreateNotificationRequest{
			UserID:       apt.OrganizerUserID,
			Type:         notification.TypeAppointmentReminder,
			Priority:     notification.PriorityNormal,
			Title:        "Upcoming Appointment",
			Message:      fmt.Sprintf("%s starts in %d minutes", apt.Title, minutesUntil),
			ResourceType: &resourceType,
			ResourceID:   &resourceID,
		})

		w.logger.Info(ctx, "worker", "Sent appointment reminder",
			zap.String("appointmentID", apt.ID),
			zap.String("title", apt.Title),
		)
	}
}

// checkEvaluationsDueSoon sends reminders for evaluations due in the next 3 days
func (w *NotificationWorker) checkEvaluationsDueSoon(ctx context.Context) {
	evaluations, err := w.store.GetEvaluationsDueSoon(ctx)
	if err != nil {
		w.logger.Error(ctx, "worker", "Failed to get evaluations due soon", zap.Error(err))
		return
	}

	for _, eval := range evaluations {
		key := fmt.Sprintf("evaluation:%s:%s", eval.ClientID, util.PgtypeDateToStr(eval.NextEvaluationDate))
		if !shouldSendNotification(key) {
			continue
		}

		resourceType := notification.ResourceTypeEvaluation
		resourceID := eval.ClientID

		// Calculate days until due
		dueDate := eval.NextEvaluationDate.Time
		daysUntil := int(time.Until(dueDate).Hours() / 24)

		urgency := notification.PriorityNormal
		if daysUntil <= 1 {
			urgency = notification.PriorityHigh
		}

		message := fmt.Sprintf("Evaluation for %s %s is due", eval.FirstName, eval.LastName)
		if daysUntil == 0 {
			message = fmt.Sprintf("Evaluation for %s %s is due today", eval.FirstName, eval.LastName)
		} else if daysUntil == 1 {
			message = fmt.Sprintf("Evaluation for %s %s is due tomorrow", eval.FirstName, eval.LastName)
		} else {
			message = fmt.Sprintf("Evaluation for %s %s is due in %d days", eval.FirstName, eval.LastName, daysUntil)
		}

		w.notificationService.Enqueue(&notification.CreateNotificationRequest{
			UserID:       eval.CoordinatorUserID,
			Type:         notification.TypeEvaluationDue,
			Priority:     urgency,
			Title:        "Evaluation Due",
			Message:      message,
			ResourceType: &resourceType,
			ResourceID:   &resourceID,
		})

		w.logger.Info(ctx, "worker", "Sent evaluation reminder",
			zap.String("clientID", eval.ClientID),
			zap.String("clientName", fmt.Sprintf("%s %s", eval.FirstName, eval.LastName)),
			zap.Int("daysUntil", daysUntil),
		)
	}
}

// checkPendingReminders sends notifications for reminders due soon
func (w *NotificationWorker) checkPendingReminders(ctx context.Context) {
	reminders, err := w.store.GetPendingRemindersByDueTime(ctx)
	if err != nil {
		w.logger.Error(ctx, "worker", "Failed to get pending reminders", zap.Error(err))
		return
	}

	for _, rem := range reminders {
		key := fmt.Sprintf("reminder:%s", rem.ID)
		if !shouldSendNotification(key) {
			continue
		}

		w.notificationService.Enqueue(&notification.CreateNotificationRequest{
			UserID:   rem.UserID,
			Type:     notification.TypeAppointmentReminder,
			Priority: notification.PriorityNormal,
			Title:    "Reminder",
			Message:  rem.Title,
		})

		w.logger.Info(ctx, "worker", "Sent reminder notification",
			zap.String("reminderID", rem.ID),
			zap.String("title", rem.Title),
		)
	}
}
