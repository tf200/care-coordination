package employee

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type employeeService struct {
	store  *db.Store
	logger *logger.Logger
}

func NewEmployeeService(store *db.Store, logger *logger.Logger) EmployeeService {
	return &employeeService{
		store:  store,
		logger: logger,
	}
}

func (s *employeeService) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (CreateEmployeeResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(ctx, "CreateEmployee", "Failed to generate password hash", zap.Error(err))
		return CreateEmployeeResponse{}, ErrInternal
	}
	id := nanoid.Generate()
	err = s.store.CreateEmployeeTx(ctx, db.CreateEmployeeTxParams{
		Emp: db.CreateEmployeeParams{
			ID:          id,
			UserID:      req.Email,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Bsn:         req.BSN,
			DateOfBirth: pgtype.Date{Time: req.DateOfBirth, Valid: true},
			PhoneNumber: req.PhoneNumber,
			Gender:      db.GenderEnum(req.Gender),
			Role:        req.Role,
		},
		User: db.CreateUserParams{
			ID:           nanoid.Generate(),
			Email:        req.Email,
			PasswordHash: string(passwordHash),
		},
	})
	if err != nil {
		s.logger.Error(ctx, "CreateEmployee", "Failed to create employee", zap.Error(err))
		return CreateEmployeeResponse{}, ErrInternal
	}
	return CreateEmployeeResponse{
		ID: id,
	}, nil
}

func (s *employeeService) ListEmployees(ctx context.Context, req *ListEmployeesRequest) (*resp.PaginationResponse[ListEmployeesResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	employees, err := s.store.ListEmployees(ctx, db.ListEmployeesParams{
		Limit:  limit,
		Offset: offset,
		Search: req.Search,
	})
	if err != nil {
		s.logger.Error(ctx, "ListEmployees", "Failed to list employees", zap.Error(err))
		return nil, ErrInternal
	}

	listEmployeesResponse := []ListEmployeesResponse{}
	totalCount := 0

	for _, employee := range employees {
		listEmployeesResponse = append(listEmployeesResponse, ListEmployeesResponse{
			ID:        employee.ID,
			FirstName: employee.FirstName,
			LastName:  employee.LastName,
		})
		if totalCount == 0 {
			totalCount = int(employee.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listEmployeesResponse, totalCount, page, pageSize)
	return &result, nil
}
