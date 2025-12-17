package rbac

// ============================================================
// Role DTOs
// ============================================================

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   *string  `json:"description"`
	PermissionIDs []string `json:"permissionIds"`
}

type CreateRoleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   *string  `json:"description"`
	PermissionIDs []string `json:"permissionIds"`
}

type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description *string              `json:"description"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
}

type RoleListItem struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	PermissionCount int64   `json:"permissionCount"`
	UserCount       int64   `json:"userCount"`
}

type ListRolesRequest struct {
	// No filters for now
}

// Permission DTOs (read-only, system-defined)
// ============================================================

type PermissionResponse struct {
	ID          string  `json:"id"`
	Resource    string  `json:"resource"`
	Action      string  `json:"action"`
	Description *string `json:"description"`
}

type ListPermissionsRequest struct {
	// No filters for now
}

// ============================================================
// Role-Permission Assignment DTOs
// ============================================================

type AssignPermissionRequest struct {
	PermissionID string `json:"permissionId" binding:"required"`
}

type RemovePermissionRequest struct {
	PermissionID string `json:"permissionId" binding:"required"`
}

// ============================================================
// User-Role Assignment DTOs
// ============================================================

type AssignRoleToUserRequest struct {
	UserID string `json:"userId" binding:"required"`
	RoleID string `json:"roleId" binding:"required"`
}

type RemoveRoleFromUserRequest struct {
	UserID string `json:"userId" binding:"required"`
	RoleID string `json:"roleId" binding:"required"`
}

type UserWithRoleResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
