package dashboard

import "context"

//go:generate mockgen -destination=../../internal/mocks/mock_dashboard_service.go -package=mocks care-cordination/features/dashboard DashboardService
type DashboardService interface {
	// Admin Dashboard
	GetOverviewStats(ctx context.Context) (*OverviewResponse, error)
	GetCriticalAlerts(ctx context.Context) (*CriticalAlertsResponse, error)
	GetPipelineStats(ctx context.Context) (*PipelineStatsResponse, error)
	GetCareTypeDistribution(ctx context.Context) (*CareTypeDistributionResponse, error)
	GetLocationCapacity(ctx context.Context, req *LocationCapacityRequest) (*LocationCapacityResponse, error)
	GetTodayAppointments(ctx context.Context, employeeID string) (*TodayAppointmentsResponse, error)
	GetEvaluationStats(ctx context.Context) (*EvaluationStatsResponse, error)
	GetDischargeStats(ctx context.Context) (*DischargeStatsResponse, error)
	// Coordinator Dashboard
	GetCoordinatorUrgentAlerts(ctx context.Context, employeeID string) (*CoordinatorUrgentAlertsResponse, error)
	GetCoordinatorTodaySchedule(ctx context.Context, employeeID string) (*CoordinatorTodayScheduleResponse, error)
	GetCoordinatorStats(ctx context.Context, employeeID string) (*CoordinatorStatsResponse, error)
	GetCoordinatorReminders(ctx context.Context, employeeID string) (*CoordinatorRemindersResponse, error)
	GetCoordinatorClients(ctx context.Context, employeeID string) (*CoordinatorClientsResponse, error)
	GetCoordinatorGoalsProgress(ctx context.Context, employeeID string) (*CoordinatorGoalsProgressResponse, error)
	GetCoordinatorIncidents(ctx context.Context, employeeID string) (*CoordinatorIncidentsResponse, error)
}
