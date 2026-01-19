package dashboard

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"context"
	"fmt"

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

func (s *dashboardService) GetOverviewStats(ctx context.Context) (*OverviewDTO, error) {
	stats, err := s.db.GetDashboardOverviewStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetOverviewStats", "Failed to get dashboard overview stats", zap.Error(err))
		return nil, ErrInternal
	}

	return &OverviewDTO{
		TotalActiveClients:   int(stats.TotalActiveClients),
		WaitingListCount:     int(stats.WaitingListCount),
		PendingRegistrations: int(stats.PendingRegistrations),
		TotalCoordinators:    int(stats.TotalCoordinators),
		TotalEmployees:       int(stats.TotalEmployees),
		OpenIncidents:        int(stats.OpenIncidents),
	}, nil
}

func (s *dashboardService) GetCriticalAlerts(ctx context.Context) (*CriticalAlertsDTO, error) {
	data, err := s.db.GetCriticalAlertsData(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetCriticalAlerts", "Failed to get critical alerts data", zap.Error(err))
		return nil, ErrInternal
	}

	alerts := []AlertDTO{}

	// Overdue evaluations - critical severity
	if data.OverdueEvaluations > 0 {
		alerts = append(alerts, AlertDTO{
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
		alerts = append(alerts, AlertDTO{
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
		alerts = append(alerts, AlertDTO{
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
		alerts = append(alerts, AlertDTO{
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
		alerts = append(alerts, AlertDTO{
			ID:          "alert-transfer",
			Type:        AlertTypeTransfer,
			Title:       fmt.Sprintf("%d verplaatsingen in afwachting", data.PendingTransfers),
			Description: "Goedkeuring vereist",
			Severity:    AlertSeverityWarning,
			Count:       int(data.PendingTransfers),
			Link:        "/verplaatsingen",
		})
	}

	return &CriticalAlertsDTO{
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

func (s *dashboardService) GetPipelineStats(ctx context.Context) (*PipelineStatsDTO, error) {
	stats, err := s.db.GetPipelineStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "GetPipelineStats", "Failed to get pipeline stats", zap.Error(err))
		return nil, ErrInternal
	}

	return &PipelineStatsDTO{
		Registrations: int(stats.Registrations),
		Intakes:       int(stats.Intakes),
		WaitingList:   int(stats.WaitingList),
		InCare:        int(stats.InCare),
		Discharged:    int(stats.Discharged),
	}, nil
}

func (s *dashboardService) GetCareTypeDistribution(ctx context.Context) (*CareTypeDistributionDTO, error) {
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
		return float64(count) / float64(total) * 100
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

	return &CareTypeDistributionDTO{
		Distribution: distribution,
		Total:        total,
	}, nil
}

func (s *dashboardService) GetLocationCapacity(ctx context.Context, req *LocationCapacityRequest) (*LocationCapacityDTO, error) {
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
			percentage = float64(occupied) / float64(capacity) * 100
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
		overallPercentage = float64(totalOccupied) / float64(totalCapacity) * 100
	}

	return &LocationCapacityDTO{
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
		sortSlice(items, func(i, j int) bool {
			return items[i].Percentage > items[j].Percentage
		})
	case "occupancy_asc":
		sortSlice(items, func(i, j int) bool {
			return items[i].Percentage < items[j].Percentage
		})
	case "name":
		sortSlice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})
	}
}

func sortSlice[T any](slice []T, less func(i, j int) bool) {
	n := len(slice)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if !less(j, j+1) {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
}

func (s *dashboardService) GetTodayAppointments(ctx context.Context, employeeID string) (*TodayAppointmentsDTO, error) {
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

	return &TodayAppointmentsDTO{
		Appointments: items,
		Count:        len(items),
	}, nil
}

func (s *dashboardService) GetEvaluationStats(ctx context.Context) (*EvaluationStatsDTO, error) {
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

	return &EvaluationStatsDTO{
		CompletionRate: completionRate,
		Completed:      completed,
		Total:          total,
		Overdue:        int(stats.Overdue),
		DueSoon:        int(stats.DueSoon),
	}, nil
}

func (s *dashboardService) GetDischargeStats(ctx context.Context) (*DischargeStatsDTO, error) {
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

	return &DischargeStatsDTO{
		ThisMonth:         int(stats.ThisMonth),
		ThisYear:          int(stats.ThisYear),
		PlannedRate:       plannedRate,
		AverageDaysInCare: int(stats.AvgDaysInCare),
	}, nil
}
