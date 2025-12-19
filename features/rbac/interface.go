package rbac

import (
	"care-cordination/lib/resp"
	"context"
)

type RBACService interface {
	// Roles
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*CreateRoleResponse, error)
	GetRole(ctx context.Context, id string) (*RoleResponse, error)
	ListRoles(
		ctx context.Context,
		req *ListRolesRequest,
	) (*resp.PaginationResponse[RoleListItem], error)
	UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error

	// Permissions (read-only, system-defined)
	ListPermissions(
		ctx context.Context,
		req *ListPermissionsRequest,
	) (*resp.PaginationResponse[PermissionResponse], error)

	// Role-Permission assignments
	AssignPermissionToRole(ctx context.Context, roleID string, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error
	ListPermissionsForRole(ctx context.Context, roleID string) ([]PermissionResponse, error)

	// User-Role assignments
	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID string) error
	GetRoleForUser(ctx context.Context, userID string) (*RoleResponse, error)
}
