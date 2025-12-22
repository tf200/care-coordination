package employee

import (
	"care-cordination/features/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
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

func (s *employeeService) CreateEmployee(
	ctx context.Context,
	req *CreateEmployeeRequest,
) (CreateEmployeeResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(ctx, "CreateEmployee", "Failed to generate password hash", zap.Error(err))
		return CreateEmployeeResponse{}, ErrInternal
	}
	id := nanoid.Generate()
	err = s.store.CreateEmployeeTx(ctx, db.CreateEmployeeTxParams{
		Emp: db.CreateEmployeeParams{
			ID:            id,
			LocationID:    req.LocationID,
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			Bsn:           req.BSN,
			DateOfBirth:   pgtype.Date{Time: req.DateOfBirth, Valid: true},
			PhoneNumber:   req.PhoneNumber,
			Gender:        db.GenderEnum(req.Gender),
			ContractHours: req.ContractHours,
			ContractType: func() db.NullContractTypeEnum {
				if req.ContractType != nil {
					return db.NullContractTypeEnum{
						Valid:            true,
						ContractTypeEnum: db.ContractTypeEnum(*req.ContractType),
					}
				}
				return db.NullContractTypeEnum{
					Valid: false,
				}
			}(),
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

func (s *employeeService) ListEmployees(
	ctx context.Context,
	req *ListEmployeesRequest,
) (*resp.PaginationResponse[ListEmployeesResponse], error) {
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
			ID:            employee.ID,
			UserID:        employee.UserID,
			FirstName:     employee.FirstName,
			LastName:      employee.LastName,
			Email:         employee.Email,
			BSN:           employee.Bsn,
			DateOfBirth:   employee.DateOfBirth.Time.Format("2006-01-02"),
			PhoneNumber:   employee.PhoneNumber,
			Gender:        string(employee.Gender),
			LocationID:    employee.LocationID,
			LocationName:  employee.LocationName,
			ContractHours: employee.ContractHours,
			ContractType: func() *string {
				if employee.ContractType.Valid {
					ct := string(employee.ContractType.ContractTypeEnum)
					return &ct
				}
				return nil
			}(),
			ClientCount: employee.ClientCount.(int64),
		})
		if totalCount == 0 {
			totalCount = int(employee.TotalCount)
		}
	}

	result := resp.PagRespWithParams(listEmployeesResponse, totalCount, page, pageSize)
	return &result, nil
}

func (s *employeeService) GetEmployeeByID(ctx context.Context, id string) (*GetEmployeeByIDResponse, error) {
	employee, err := s.store.GetEmployeeByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "GetEmployeeByID", "Failed to get employee", zap.Error(err))
		return nil, ErrInternal
	}

	return &GetEmployeeByIDResponse{
		ID:            employee.ID,
		UserID:        employee.UserID,
		FirstName:     employee.FirstName,
		LastName:      employee.LastName,
		Email:         employee.Email,
		BSN:           employee.Bsn,
		DateOfBirth:   employee.DateOfBirth.Time.Format("2006-01-02"),
		PhoneNumber:   employee.PhoneNumber,
		Gender:        string(employee.Gender),
		LocationID:    employee.LocationID,
		LocationName:  employee.LocationName,
		ContractHours: employee.ContractHours,
		ContractType: func() *string {
			if employee.ContractType.Valid {
				ct := string(employee.ContractType.ContractTypeEnum)
				return &ct
			}
			return nil
		}(),
		RoleID:      employee.RoleID,
		RoleName:    employee.RoleName,
		ClientCount: employee.ClientCount.(int64),
	}, nil
}

func (s *employeeService) GetMyProfile(ctx context.Context) (*GetMyProfileResponse, error) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return nil, ErrUnauthorized
	}

	employee, err := s.store.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "GetMyProfile", "Failed to get employee profile", zap.Error(err))
		return nil, ErrInternal
	}

	// Role name comes from LEFT JOIN in the query
	var roleName string
	if employee.RoleName != nil {
		roleName = *employee.RoleName
	}

	// Fetch permissions for the user's role
	permissionsResponse := []PermissionResponse{}
	if employee.RoleID != nil {
		permissions, err := s.store.ListPermissionsForRole(ctx, *employee.RoleID)
		if err != nil {
			s.logger.Error(ctx, "GetMyProfile", "Failed to get role permissions", zap.Error(err))
			return nil, ErrInternal
		}

		for _, perm := range permissions {
			desc := ""
			if perm.Description != nil {
				desc = *perm.Description
			}
			permissionsResponse = append(permissionsResponse, PermissionResponse{
				ID:          perm.ID,
				Resource:    perm.Resource,
				Action:      perm.Action,
				Description: desc,
			})
		}
	}

	return &GetMyProfileResponse{
		ID:          employee.ID,
		UserID:      employee.UserID,
		FirstName:   employee.FirstName,
		LastName:    employee.LastName,
		Email:       employee.Email,
		BSN:         employee.Bsn,
		DateOfBirth: employee.DateOfBirth.Time.Format("2006-01-02"),
		PhoneNumber: employee.PhoneNumber,
		Gender:      string(employee.Gender),
		Role:        roleName,
		Permissions: permissionsResponse,
	}, nil
}
