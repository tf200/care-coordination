package dashboard

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/util"
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
)

type dashboardService struct {
	db     db.StoreInterface
	logger logger.Logger
}

func NewDashboardService(
	db db.StoreInterface,
	logger logger.Logger,
) DashboardService {
	return &dashboardService{
		db:     db,
		logger: logger,
	}
}

func (s *dashboardService) GetOverviewStats(ctx context.Context) (*OverviewResponse, error) {
	stats, err := s.db.GetDashboardOverviewStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetOverviewStats", "Failed to get dashboard overview stats", zap.Error(err))
		return nil, ErrInternal
	}

	return &OverviewResponse{
		TotalActiveClients:   int(stats.TotalActiveClients),
		WaitingListCount:     int(stats.WaitingListCount),
		PendingRegistrations: int(stats.PendingRegistrations),
		TotalCoordinators:    int(stats.TotalCoordinators),
		TotalEmployees:       int(stats.TotalEmployees),
		OpenIncidents:        int(stats.OpenIncidents),
	}, nil
}

func (s *dashboardService) GetCriticalAlerts(ctx context.Context) (*CriticalAlertsResponse, error) {
	data, err := s.db.GetCriticalAlertsData(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetCriticalAlerts", "Failed to get critical alerts data", zap.Error(err))
		return nil, ErrInternal
	}

	alerts := []AlertItem{}

	// Overdue evaluations - critical severity
	if data.OverdueEvaluations > 0 {
		alerts = append(alerts, AlertItem{
			ID:          "alert-evaluation",
			Type:        AlertTypeEvaluation,
			Title:       fmt.Sprintf("%d evaluaties achterstallig", data.OverdueEvaluations),
			Description: "Vereist onmiddellijke actie",
			Severity:    AlertSeverityCritical,
			Count:       int(data.OverdueEvaluations),
			Link:        "/evaluaties",
		})
	}

	// Care end date approaching - warning severity
	if data.CareEndingSoon > 0 {
		alerts = append(alerts, AlertItem{
			ID:          "alert-care-end",
			Type:        AlertTypeCareEnd,
			Title:       fmt.Sprintf("%d zorgtrajecten eindigen binnen 30 dagen", data.CareEndingSoon),
			Description: "Herindicatie nodig",
			Severity:    AlertSeverityWarning,
			Count:       int(data.CareEndingSoon),
			Link:        "/inzorg",
		})
	}

	// Open incidents - severity based on incident severity counts
	if data.OpenIncidents > 0 {
		severity := AlertSeverityWarning
		if data.SevereIncidents > 0 {
			severity = AlertSeverityCritical
		}

		description := s.buildIncidentDescription(int(data.SevereIncidents), int(data.ModerateIncidents))
		alerts = append(alerts, AlertItem{
			ID:          "alert-incident",
			Type:        AlertTypeIncident,
			Title:       fmt.Sprintf("%d incidenten in onderzoek", data.OpenIncidents),
			Description: description,
			Severity:    severity,
			Count:       int(data.OpenIncidents),
			Link:        "/incidenten",
		})
	}

	// High priority waiting list - warning severity
	if data.HighPriorityWaiting > 0 {
		alerts = append(alerts, AlertItem{
			ID:          "alert-waitlist",
			Type:        AlertTypeWaitlist,
			Title:       fmt.Sprintf("%d hoge prioriteit op wachtlijst", data.HighPriorityWaiting),
			Description: "Wacht op plaatsing",
			Severity:    AlertSeverityWarning,
			Count:       int(data.HighPriorityWaiting),
			Link:        "/wachtlijst",
		})
	}

	// Pending location transfers - warning severity
	if data.PendingTransfers > 0 {
		alerts = append(alerts, AlertItem{
			ID:          "alert-transfer",
			Type:        AlertTypeTransfer,
			Title:       fmt.Sprintf("%d verplaatsingen in afwachting", data.PendingTransfers),
			Description: "Goedkeuring vereist",
			Severity:    AlertSeverityWarning,
			Count:       int(data.PendingTransfers),
			Link:        "/verplaatsingen",
		})
	}

	return &CriticalAlertsResponse{
		Alerts: alerts,
	}, nil
}

func (s *dashboardService) buildIncidentDescription(severe, moderate int) string {
	parts := []string{}
	if severe > 0 {
		parts = append(parts, fmt.Sprintf("%d ernstig", severe))
	}
	if moderate > 0 {
		parts = append(parts, fmt.Sprintf("%d gemiddeld", moderate))
	}
	if len(parts) == 0 {
		return "In onderzoek"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + ", " + parts[1]
}

func (s *dashboardService) GetPipelineStats(ctx context.Context) (*PipelineStatsResponse, error) {
	stats, err := s.db.GetPipelineStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetPipelineStats", "Failed to get pipeline stats", zap.Error(err))
		return nil, ErrInternal
	}

	return &PipelineStatsResponse{
		Registrations: int(stats.Registrations),
		Intakes:       int(stats.Intakes),
		WaitingList:   int(stats.WaitingList),
		InCare:        int(stats.InCare),
		Discharged:    int(stats.Discharged),
	}, nil
}

func (s *dashboardService) GetCareTypeDistribution(ctx context.Context) (*CareTypeDistributionResponse, error) {
	data, err := s.db.GetCareTypeDistribution(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetCareTypeDistribution", "Failed to get care type distribution", zap.Error(err))
		return nil, ErrInternal
	}

	total := int(data.Total)
	distribution := []CareTypeDistributionItem{}

	// Helper to calculate percentage
	calcPercentage := func(count int64) float64 {
		if total == 0 {
			return 0
		}
		val := float64(count) / float64(total) * 100
		return math.Round(val*100) / 100
	}

	// Add care types with their Dutch labels
	if data.ProtectedLiving > 0 {
		distribution = append(distribution, CareTypeDistributionItem{
			CareType:   "protected_living",
			Label:      "Beschermd Wonen",
			Count:      int(data.ProtectedLiving),
			Percentage: calcPercentage(data.ProtectedLiving),
		})
	}

	if data.SemiIndependentLiving > 0 {
		distribution = append(distribution, CareTypeDistributionItem{
			CareType:   "semi_independent_living",
			Label:      "Semi-zelfstandig Wonen",
			Count:      int(data.SemiIndependentLiving),
			Percentage: calcPercentage(data.SemiIndependentLiving),
		})
	}

	if data.IndependentAssistedLiving > 0 {
		distribution = append(distribution, CareTypeDistributionItem{
			CareType:   "independent_assisted_living",
			Label:      "Begeleid Zelfstandig Wonen",
			Count:      int(data.IndependentAssistedLiving),
			Percentage: calcPercentage(data.IndependentAssistedLiving),
		})
	}

	if data.AmbulatoryCare > 0 {
		distribution = append(distribution, CareTypeDistributionItem{
			CareType:   "ambulatory_care",
			Label:      "Ambulante Zorg",
			Count:      int(data.AmbulatoryCare),
			Percentage: calcPercentage(data.AmbulatoryCare),
		})
	}

	return &CareTypeDistributionResponse{
		Distribution: distribution,
		Total:        total,
	}, nil
}

func (s *dashboardService) GetLocationCapacity(ctx context.Context, req *LocationCapacityRequest) (*LocationCapacityResponse, error) {
	// Get all locations
	locations, err := s.db.GetLocationCapacityList(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetLocationCapacity", "Failed to get location capacity list", zap.Error(err))
		return nil, ErrInternal
	}

	// Get totals
	totals, err := s.db.GetLocationCapacityTotals(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetLocationCapacity", "Failed to get location capacity totals", zap.Error(err))
		return nil, ErrInternal
	}

	// Convert to DTOs
	items := make([]LocationCapacityItem, len(locations))
	for i, loc := range locations {
		capacity := int(loc.Capacity)
		occupied := int(loc.Occupied)
		available := capacity - occupied
		percentage := float64(0)
		if capacity > 0 {
			val := float64(occupied) / float64(capacity) * 100
			percentage = math.Round(val*100) / 100
		}
		items[i] = LocationCapacityItem{
			ID:         loc.ID,
			Name:       loc.Name,
			Capacity:   capacity,
			Occupied:   occupied,
			Available:  available,
			Percentage: percentage,
		}
	}

	// Sort based on request
	sort := req.Sort
	if sort == "" {
		sort = "occupancy_desc"
	}
	s.sortLocationItems(items, sort)

	// Apply limit
	limit := req.Limit
	if limit <= 0 {
		limit = 4
	}
	if limit > len(items) {
		limit = len(items)
	}
	items = items[:limit]

	// Calculate totals
	totalCapacity := int(totals.TotalCapacity)
	totalOccupied := int(totals.TotalOccupied)
	totalAvailable := totalCapacity - totalOccupied
	overallPercentage := float64(0)
	if totalCapacity > 0 {
		val := float64(totalOccupied) / float64(totalCapacity) * 100
		overallPercentage = math.Round(val*100) / 100
	}

	return &LocationCapacityResponse{
		Locations: items,
		Totals: LocationCapacityTotals{
			TotalCapacity:     totalCapacity,
			TotalOccupied:     totalOccupied,
			TotalAvailable:    totalAvailable,
			OverallPercentage: overallPercentage,
		},
	}, nil
}

func (s *dashboardService) sortLocationItems(items []LocationCapacityItem, sortBy string) {
	switch sortBy {
	case "occupancy_desc":
		util.SortSlice(items, func(i, j int) bool {
			return items[i].Percentage > items[j].Percentage
		})
	case "occupancy_asc":
		util.SortSlice(items, func(i, j int) bool {
			return items[i].Percentage < items[j].Percentage
		})
	case "name":
		util.SortSlice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})
	}
}

func (s *dashboardService) GetTodayAppointments(ctx context.Context, employeeID string) (*TodayAppointmentsResponse, error) {
	appointments, err := s.db.GetTodayAppointmentsForEmployee(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetTodayAppointments", "Failed to get today's appointments", zap.Error(err))
		return nil, ErrInternal
	}

	items := make([]TodayAppointmentItem, len(appointments))
	for i, apt := range appointments {
		locationName := ""
		if apt.Location != nil {
			locationName = *apt.Location
		}
		items[i] = TodayAppointmentItem{
			ID:           apt.ID,
			Type:         string(apt.Type),
			Title:        apt.Title,
			ClientID:     apt.ClientID,
			ClientName:   apt.ClientName,
			StartTime:    apt.StartTime.Time.Format("15:04"),
			EndTime:      apt.EndTime.Time.Format("15:04"),
			LocationName: locationName,
		}
	}

	return &TodayAppointmentsResponse{
		Appointments: items,
		Count:        len(items),
	}, nil
}

func (s *dashboardService) GetEvaluationStats(ctx context.Context) (*EvaluationStatsResponse, error) {
	stats, err := s.db.GetEvaluationStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetEvaluationStats", "Failed to get evaluation stats", zap.Error(err))
		return nil, ErrInternal
	}

	total := int(stats.Total)
	completed := int(stats.Completed)
	completionRate := 0
	if total > 0 {
		completionRate = (completed * 100) / total
	}

	return &EvaluationStatsResponse{
		CompletionRate: completionRate,
		Completed:      completed,
		Total:          total,
		Overdue:        int(stats.Overdue),
		DueSoon:        int(stats.DueSoon),
	}, nil
}

func (s *dashboardService) GetDischargeStats(ctx context.Context) (*DischargeStatsResponse, error) {
	stats, err := s.db.GetDashboardDischargeStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetDischargeStats", "Failed to get discharge stats", zap.Error(err))
		return nil, ErrInternal
	}

	totalDischarged := int(stats.TotalDischarged)
	plannedDischarges := int(stats.PlannedDischarges)
	plannedRate := 0
	if totalDischarged > 0 {
		plannedRate = (plannedDischarges * 100) / totalDischarged
	}

	return &DischargeStatsResponse{
		ThisMonth:         int(stats.ThisMonth),
		ThisYear:          int(stats.ThisYear),
		PlannedRate:       plannedRate,
		AverageDaysInCare: int(stats.AvgDaysInCare),
	}, nil
}

// Coordinator Dashboard Methods

func (s *dashboardService) GetCoordinatorUrgentAlerts(ctx context.Context, employeeID string) (*CoordinatorUrgentAlertsResponse, error) {
	// Get counts
	data, err := s.db.GetCoordinatorUrgentAlertsData(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorUrgentAlerts", "Failed to get coordinator urgent alerts data", zap.Error(err))
		return nil, ErrInternal
	}

	alerts := []CoordinatorUrgentAlertItem{}

	// Overdue evaluations (critical)
	if data.OverdueEvaluations > 0 {
		clients, _ := s.db.GetCoordinatorOverdueEvaluationClients(ctx, employeeID)
		clientIDs, description := s.buildClientInfo(clients)
		alerts = append(alerts, CoordinatorUrgentAlertItem{
			ID:          "alert-evaluation",
			Type:        CoordinatorAlertTypeEvaluation,
			Title:       "overdue_evaluations",
			Description: description,
			Severity:    CoordinatorAlertSeverityCritical,
			Count:       int(data.OverdueEvaluations),
			ClientIDs:   clientIDs,
			Link:        "/evaluaties",
		})
	}

	// Expiring contracts (warning)
	if data.ExpiringContracts > 0 {
		clients, _ := s.db.GetCoordinatorExpiringContractClients(ctx, employeeID)
		clientIDs, description := s.buildClientInfo(clients)
		alerts = append(alerts, CoordinatorUrgentAlertItem{
			ID:          "alert-contract",
			Type:        CoordinatorAlertTypeContract,
			Title:       "expiring_contracts",
			Description: description,
			Severity:    CoordinatorAlertSeverityWarning,
			Count:       int(data.ExpiringContracts),
			ClientIDs:   clientIDs,
			Link:        "/inzorg",
		})
	}

	// Draft evaluations (info)
	if data.DraftEvaluations > 0 {
		clients, _ := s.db.GetCoordinatorDraftEvaluationClients(ctx, employeeID)
		clientIDs, description := s.buildClientInfo(clients)
		alerts = append(alerts, CoordinatorUrgentAlertItem{
			ID:          "alert-draft",
			Type:        CoordinatorAlertTypeDraft,
			Title:       "draft_evaluations",
			Description: description,
			Severity:    CoordinatorAlertSeverityInfo,
			Count:       int(data.DraftEvaluations),
			ClientIDs:   clientIDs,
			Link:        "/evaluaties",
		})
	}

	// Unresolved incidents (warning)
	if data.UnresolvedIncidents > 0 {
		clients, _ := s.db.GetCoordinatorUnresolvedIncidentClients(ctx, employeeID)
		clientIDs, description := s.buildClientInfo(clients)
		alerts = append(alerts, CoordinatorUrgentAlertItem{
			ID:          "alert-incident",
			Type:        CoordinatorAlertTypeIncident,
			Title:       "unresolved_incidents",
			Description: description,
			Severity:    CoordinatorAlertSeverityWarning,
			Count:       int(data.UnresolvedIncidents),
			ClientIDs:   clientIDs,
			Link:        "/incidenten",
		})
	}

	// Long waiting clients (warning)
	if data.LongWaiting > 0 {
		clients, _ := s.db.GetCoordinatorLongWaitingClients(ctx, employeeID)
		clientIDs, description := s.buildClientInfo(clients)
		alerts = append(alerts, CoordinatorUrgentAlertItem{
			ID:          "alert-waiting",
			Type:        CoordinatorAlertTypeWaitingLong,
			Title:       "long_waiting_clients",
			Description: description,
			Severity:    CoordinatorAlertSeverityWarning,
			Count:       int(data.LongWaiting),
			ClientIDs:   clientIDs,
			Link:        "/wachtlijst",
		})
	}

	return &CoordinatorUrgentAlertsResponse{Alerts: alerts}, nil
}

type clientInfo interface {
	GetID() string
	GetFirstName() string
	GetLastName() string
}

func (s *dashboardService) buildClientInfo(clients any) ([]string, string) {
	clientIDs := []string{}
	names := []string{}

	switch v := clients.(type) {
	case []db.GetCoordinatorOverdueEvaluationClientsRow:
		for _, c := range v {
			clientIDs = append(clientIDs, c.ID)
			names = append(names, c.FirstName+" "+c.LastName)
		}
	case []db.GetCoordinatorExpiringContractClientsRow:
		for _, c := range v {
			clientIDs = append(clientIDs, c.ID)
			names = append(names, c.FirstName+" "+c.LastName)
		}
	case []db.GetCoordinatorDraftEvaluationClientsRow:
		for _, c := range v {
			clientIDs = append(clientIDs, c.ID)
			names = append(names, c.FirstName+" "+c.LastName)
		}
	case []db.GetCoordinatorUnresolvedIncidentClientsRow:
		for _, c := range v {
			clientIDs = append(clientIDs, c.ID)
			names = append(names, c.FirstName+" "+c.LastName)
		}
	case []db.GetCoordinatorLongWaitingClientsRow:
		for _, c := range v {
			clientIDs = append(clientIDs, c.ID)
			names = append(names, c.FirstName+" "+c.LastName)
		}
	}

	description := ""
	if len(names) > 0 {
		if len(names) <= 3 {
			description = strings.Join(names, ", ")
		} else {
			description = strings.Join(names[:3], ", ") + fmt.Sprintf(" +%d", len(names)-3)
		}
	}

	return clientIDs, description
}

func (s *dashboardService) GetCoordinatorTodaySchedule(ctx context.Context, employeeID string) (*CoordinatorTodayScheduleResponse, error) {
	appointments, err := s.db.GetCoordinatorTodaySchedule(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorTodaySchedule", "Failed to get coordinator today schedule", zap.Error(err))
		return nil, ErrInternal
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	items := make([]CoordinatorScheduleItem, len(appointments))

	for i, apt := range appointments {
		// Determine status based on time
		status := "upcoming"
		if apt.Status.Valid {
			status = string(apt.Status.AppointmentStatusEnum)
		}
		if apt.EndTime.Time.Before(now) {
			status = "completed"
		} else if apt.StartTime.Time.Before(now) && apt.EndTime.Time.After(now) {
			status = "in_progress"
		}

		items[i] = CoordinatorScheduleItem{
			ID:           apt.ID,
			Time:         apt.StartTime.Time.Format("15:04"),
			EndTime:      apt.EndTime.Time.Format("15:04"),
			Type:         string(apt.Type),
			ClientID:     apt.ClientID,
			ClientName:   apt.ClientName,
			LocationID:   apt.LocationID,
			LocationName: apt.LocationName,
			Status:       status,
		}
	}

	return &CoordinatorTodayScheduleResponse{
		Date:         today,
		Appointments: items,
		Count:        len(items),
	}, nil
}

func (s *dashboardService) GetCoordinatorStats(ctx context.Context, employeeID string) (*CoordinatorStatsResponse, error) {
	stats, err := s.db.GetCoordinatorStats(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorStats", "Failed to get coordinator stats", zap.Error(err))
		return nil, ErrInternal
	}

	return &CoordinatorStatsResponse{
		MyActiveClients:       int(stats.MyActiveClients),
		MyUpcomingEvaluations: int(stats.MyUpcomingEvaluations),
		MyPendingIntakes:      int(stats.MyPendingIntakes),
		MyWaitingListClients:  int(stats.MyWaitingListClients),
	}, nil
}

func (s *dashboardService) GetCoordinatorReminders(ctx context.Context, employeeID string) (*CoordinatorRemindersResponse, error) {
	reminders, err := s.db.GetCoordinatorReminders(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorReminders", "Failed to get coordinator reminders", zap.Error(err))
		return nil, ErrInternal
	}

	items := make([]ReminderItem, len(reminders))
	for i, r := range reminders {
		items[i] = ReminderItem{
			ID:       r.ID,
			Title:    r.Title,
			Client:   "",
			DueDate:  r.DueTime.Time.Format("2006-01-02"),
			Priority: "medium",
		}
	}

	return &CoordinatorRemindersResponse{Reminders: items}, nil
}

func (s *dashboardService) GetCoordinatorClients(ctx context.Context, employeeID string) (*CoordinatorClientsResponse, error) {
	clients, err := s.db.GetCoordinatorClients(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorClients", "Failed to get coordinator clients", zap.Error(err))
		return nil, ErrInternal
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	items := make([]CoordinatorClientItem, len(clients))
	for i, c := range clients {
		daysUntilEnd := 0
		status := "active"
		if c.CareEndDate.Valid {
			endDate := c.CareEndDate.Time
			daysUntilEnd = int(endDate.Sub(today).Hours() / 24)
			if daysUntilEnd <= 30 {
				status = "expiring"
			}
		}

		nextEvaluation := "-"
		if c.NextEvaluationDate.Valid {
			nextEvaluation = c.NextEvaluationDate.Time.Format("2006-01-02")
		}
		evalStatus := s.calculateEvaluationStatus(c.NextEvaluationDate.Time, c.NextEvaluationDate.Valid, today)

		locationName := ""
		if c.LocationName != nil {
			locationName = *c.LocationName
		}

		items[i] = CoordinatorClientItem{
			ID:               c.ID,
			FirstName:        c.FirstName,
			LastName:         c.LastName,
			CareType:         string(c.CareType),
			Location:         locationName,
			DaysUntilEnd:     daysUntilEnd,
			Status:           status,
			NextEvaluation:   nextEvaluation,
			EvaluationStatus: evalStatus,
		}
	}

	return &CoordinatorClientsResponse{Clients: items}, nil
}

func (s *dashboardService) calculateEvaluationStatus(evalDate time.Time, valid bool, today time.Time) string {
	if !valid {
		return "upcoming"
	}

	evalDay := time.Date(evalDate.Year(), evalDate.Month(), evalDate.Day(), 0, 0, 0, 0, evalDate.Location())
	daysUntil := int(evalDay.Sub(today).Hours() / 24)

	if daysUntil <= 0 {
		return "overdue"
	} else if daysUntil <= 7 {
		return "due_soon"
	}
	return "upcoming"
}

func (s *dashboardService) GetCoordinatorGoalsProgress(ctx context.Context, employeeID string) (*CoordinatorGoalsProgressResponse, error) {
	progress, err := s.db.GetCoordinatorGoalsProgress(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorGoalsProgress", "Failed to get coordinator goals progress", zap.Error(err))
		return nil, ErrInternal
	}

	return &CoordinatorGoalsProgressResponse{
		OnTrack:    int(progress.OnTrack),
		Delayed:    int(progress.Delayed),
		Achieved:   int(progress.Achieved),
		NotStarted: int(progress.NotStarted),
		Total:      int(progress.Total),
	}, nil
}

func (s *dashboardService) GetCoordinatorIncidents(ctx context.Context, employeeID string) (*CoordinatorIncidentsResponse, error) {
	incidents, err := s.db.GetCoordinatorIncidents(ctx, employeeID)
	if err != nil {
		s.logger.Error(ctx, "GetCoordinatorIncidents", "Failed to get coordinator incidents", zap.Error(err))
		return nil, ErrInternal
	}

	items := make([]CoordinatorIncidentItem, len(incidents))
	for i, inc := range incidents {
		items[i] = CoordinatorIncidentItem{
			ID:       inc.ID,
			Client:   inc.ClientFirstName + " " + inc.ClientLastName,
			Type:     string(inc.IncidentType),
			Severity: string(inc.IncidentSeverity),
			Date:     inc.IncidentDate.Time.Format("2006-01-02"),
			Status:   string(inc.Status),
		}
	}

	return &CoordinatorIncidentsResponse{Incidents: items}, nil
}
