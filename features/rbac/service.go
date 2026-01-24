package rbac

import (
	"care-cordination/lib/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"context"

	"go.uber.org/zap"
)

type rbacService struct {
	store  *db.Store
	logger logger.Logger
}

func NewRBACService(store *db.Store, logger logger.Logger) RBACService {
	return &rbacService{
		store:  store,
		logger: logger,
	}
}

// ============================================================
// Roles
// ============================================================

func (s *rbacService) CreateRole(
	ctx context.Context,
	req *CreateRoleRequest,
) (*CreateRoleResponse, error) {
	id := nanoid.Generate()

	result, err := s.store.CreateRoleWithPermissionsTx(ctx, db.CreateRoleWithPermissionsTxParams{
		Role: db.CreateRoleParams{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
		},
		PermissionIDs: req.PermissionIDs,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrRoleAlreadyExists
		}
		s.logger.Error(ctx, "CreateRole", "Failed to create role", zap.Error(err))
		return nil, ErrInternal
	}
	return &CreateRoleResponse{
		ID:          result.Role.ID,
		Name:        result.Role.Name,
		Description: result.Role.Description,
	}, nil
}

func (s *rbacService) GetRole(ctx context.Context, id string) (*RoleResponse, error) {
	role, err := s.store.GetRoleByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "GetRole", "Failed to get role", zap.Error(err))
		return nil, ErrRoleNotFound
	}

	// Fetch permissions for this role
	permissions, err := s.store.ListPermissionsForRole(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "GetRole", "Failed to get permissions for role", zap.Error(err))
		return nil, ErrInternal
	}

	permissionResponses := make([]PermissionResponse, len(permissions))
	for i, perm := range permissions {
		permissionResponses[i] = PermissionResponse{
			ID:          perm.ID,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Description: perm.Description,
		}
	}

	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissionResponses,
	}, nil
}

func (s *rbacService) ListRoles(
	ctx context.Context,
	req *ListRolesRequest,
) (*resp.PaginationResponse[RoleListItem], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	roles, err := s.store.ListRoles(ctx, db.ListRolesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.logger.Error(ctx, "ListRoles", "Failed to list roles", zap.Error(err))
		return nil, ErrInternal
	}

	roleResponses := []RoleListItem{}
	totalCount := 0

	for _, role := range roles {
		roleResponses = append(roleResponses, RoleListItem{
			ID:              role.ID,
			Name:            role.Name,
			Description:     role.Description,
			PermissionCount: role.PermissionCount,
			UserCount:       role.UserCount,
		})
		if totalCount == 0 {
			totalCount = int(role.TotalCount)
		}
	}

	result := resp.PagRespWithParams(roleResponses, totalCount, page, pageSize)
	return &result, nil
}

func (s *rbacService) UpdateRole(
	ctx context.Context,
	id string,
	req *UpdateRoleRequest,
) (*RoleResponse, error) {
	result, err := s.store.UpdateRoleWithPermissionsTx(ctx, db.UpdateRoleWithPermissionsTxParams{
		Role: db.UpdateRoleParams{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
		},
		PermissionIDs: req.PermissionIDs,
	})
	if err != nil {
		s.logger.Error(ctx, "UpdateRole", "Failed to update role", zap.Error(err))
		return nil, ErrInternal
	}

	// Fetch the updated permissions
	permissions, err := s.store.ListPermissionsForRole(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "UpdateRole", "Failed to get permissions for role", zap.Error(err))
		return nil, ErrInternal
	}

	permissionResponses := make([]PermissionResponse, len(permissions))
	for i, perm := range permissions {
		permissionResponses[i] = PermissionResponse{
			ID:          perm.ID,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Description: perm.Description,
		}
	}

	return &RoleResponse{
		ID:          result.Role.ID,
		Name:        result.Role.Name,
		Description: result.Role.Description,
		Permissions: permissionResponses,
	}, nil
}

func (s *rbacService) DeleteRole(ctx context.Context, id string) error {
	err := s.store.DeleteRole(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "DeleteRole", "Failed to delete role", zap.Error(err))
		return ErrInternal
	}
	return nil
}

// ============================================================
// Permissions (read-only, system-defined)
// ============================================================

func (s *rbacService) ListPermissions(
	ctx context.Context,
	req *ListPermissionsRequest,
) (*resp.PaginationResponse[PermissionResponse], error) {
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	permissions, err := s.store.ListPermissions(ctx, db.ListPermissionsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.logger.Error(ctx, "ListPermissions", "Failed to list permissions", zap.Error(err))
		return nil, ErrInternal
	}

	permissionResponses := []PermissionResponse{}
	totalCount := 0

	for _, perm := range permissions {
		permissionResponses = append(permissionResponses, PermissionResponse{
			ID:          perm.ID,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Description: perm.Description,
		})
		if totalCount == 0 {
			totalCount = int(perm.TotalCount)
		}
	}

	result := resp.PagRespWithParams(permissionResponses, totalCount, page, pageSize)
	return &result, nil
}

// ============================================================
// Role-Permission Assignments
// ============================================================

func (s *rbacService) AssignPermissionToRole(
	ctx context.Context,
	roleID string,
	permissionID string,
) error {
	err := s.store.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	if err != nil {
		s.logger.Error(ctx, "AssignPermissionToRole", "Failed to assign permission", zap.Error(err))
		return ErrInternal
	}
	return nil
}

func (s *rbacService) RemovePermissionFromRole(
	ctx context.Context,
	roleID string,
	permissionID string,
) error {
	err := s.store.RemovePermissionFromRole(ctx, db.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	if err != nil {
		s.logger.Error(
			ctx,
			"RemovePermissionFromRole",
			"Failed to remove permission",
			zap.Error(err),
		)
		return ErrInternal
	}
	return nil
}

func (s *rbacService) ListPermissionsForRole(
	ctx context.Context,
	roleID string,
) ([]PermissionResponse, error) {
	permissions, err := s.store.ListPermissionsForRole(ctx, roleID)
	if err != nil {
		s.logger.Error(ctx, "ListPermissionsForRole", "Failed to list permissions", zap.Error(err))
		return nil, ErrInternal
	}

	result := []PermissionResponse{}
	for _, perm := range permissions {
		result = append(result, PermissionResponse{
			ID:          perm.ID,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Description: perm.Description,
		})
	}
	return result, nil
}

// ============================================================
// User-Role Assignments
// ============================================================

func (s *rbacService) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	err := s.store.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		s.logger.Error(ctx, "AssignRoleToUser", "Failed to assign role", zap.Error(err))
		return ErrInternal
	}
	return nil
}

func (s *rbacService) RemoveRoleFromUser(ctx context.Context, userID string) error {
	err := s.store.RemoveRoleFromUser(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "RemoveRoleFromUser", "Failed to remove role", zap.Error(err))
		return ErrInternal
	}
	return nil
}

func (s *rbacService) GetRoleForUser(ctx context.Context, userID string) (*RoleResponse, error) {
	role, err := s.store.GetRoleForUser(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "GetRoleForUser", "Failed to get role", zap.Error(err))
		return nil, ErrInternal
	}

	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}, nil
}
