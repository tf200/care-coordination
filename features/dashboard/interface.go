package dashboard

import "context"

//go:generate mockgen -destination=../../internal/mocks/mock_dashboard_service.go -package=mocks care-cordination/features/dashboard DashboardService
type DashboardService interface {
	// Admin Dashboard
	GetOverviewStats(ctx context.Context) (*OverviewDTO, error)
	GetCriticalAlerts(ctx context.Context) (*CriticalAlertsDTO, error)
	GetPipelineStats(ctx context.Context) (*PipelineStatsDTO, error)
	GetCareTypeDistribution(ctx context.Context) (*CareTypeDistributionDTO, error)
	GetLocationCapacity(ctx context.Context, req *LocationCapacityRequest) (*LocationCapacityDTO, error)
	GetTodayAppointments(ctx context.Context, employeeID string) (*TodayAppointmentsDTO, error)
	GetEvaluationStats(ctx context.Context) (*EvaluationStatsDTO, error)
	GetDischargeStats(ctx context.Context) (*DischargeStatsDTO, error)
	// Coordinator Dashboard
	GetCoordinatorUrgentAlerts(ctx context.Context, employeeID string) (*CoordinatorUrgentAlertsDTO, error)
	GetCoordinatorTodaySchedule(ctx context.Context, employeeID string) (*CoordinatorTodayScheduleDTO, error)
	GetCoordinatorStats(ctx context.Context, employeeID string) (*CoordinatorStatsDTO, error)
	GetCoordinatorReminders(ctx context.Context, employeeID string) (*CoordinatorRemindersDTO, error)
	GetCoordinatorClients(ctx context.Context, employeeID string) (*CoordinatorClientsDTO, error)
	GetCoordinatorGoalsProgress(ctx context.Context, employeeID string) (*CoordinatorGoalsProgressDTO, error)
	GetCoordinatorIncidents(ctx context.Context, employeeID string) (*CoordinatorIncidentsDTO, error)
}
