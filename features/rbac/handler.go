package rbac

import (
	"care-cordination/lib/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	rbacService RBACService
	mdw         *middleware.Middleware
}

func NewRBACHandler(rbacService RBACService, mdw *middleware.Middleware) *RBACHandler {
	return &RBACHandler{
		rbacService: rbacService,
		mdw:         mdw,
	}
}

func (h *RBACHandler) SetupRBACRoutes(router *gin.Engine) {
	admin := router.Group("/admin")
	admin.Use(h.mdw.AuthMdw())

	// Roles
	roles := admin.Group("/roles")
	roles.POST("", h.mdw.RequirePermission("rbac", "write"), h.CreateRole)
	roles.GET("", h.mdw.PaginationMdw(), h.mdw.RequirePermission("rbac", "read"), h.ListRoles)
	roles.GET("/:id", h.mdw.RequirePermission("rbac", "read"), h.GetRole)
	roles.PUT("/:id", h.mdw.RequirePermission("rbac", "write"), h.UpdateRole)
	roles.DELETE("/:id", h.mdw.RequirePermission("rbac", "delete"), h.DeleteRole)
	roles.GET("/:id/permissions", h.mdw.RequirePermission("rbac", "read"), h.ListPermissionsForRole)
	roles.POST("/:id/permissions", h.mdw.RequirePermission("rbac", "write"), h.AssignPermissionToRole)
	roles.DELETE("/:id/permissions/:permissionId", h.mdw.RequirePermission("rbac", "delete"), h.RemovePermissionFromRole)

	// Permissions (read-only)
	permissions := admin.Group("/permissions")
	permissions.GET("", h.mdw.PaginationMdw(), h.mdw.RequirePermission("rbac", "read"), h.ListPermissions)

	// User-Role assignments
	userRoles := admin.Group("/user-roles")
	userRoles.POST("", h.mdw.RequirePermission("rbac", "write"), h.AssignRoleToUser)
	userRoles.DELETE("/user/:userId", h.mdw.RequirePermission("rbac", "delete"), h.RemoveRoleFromUser)
	userRoles.GET("/user/:userId", h.mdw.RequirePermission("rbac", "read"), h.GetRoleForUser)
}

// ============================================================
// Role Handlers
// ============================================================

// @Summary Create a role
// @Description Create a new role with optional permissions
// @Tags RBAC - Roles
// @Accept json
// @Produce json
// @Param role body CreateRoleRequest true "Role data"
// @Success 201 {object} resp.SuccessResponse[CreateRoleResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 409 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles [post]
func (h *RBACHandler) CreateRole(ctx *gin.Context) {
	var req CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.rbacService.CreateRole(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoleAlreadyExists):
			ctx.JSON(http.StatusConflict, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusCreated, resp.Success(result, "Role created successfully"))
}

// @Summary Get a role
// @Description Get a role by ID with its permissions
// @Tags RBAC - Roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} resp.SuccessResponse[RoleResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles/{id} [get]
func (h *RBACHandler) GetRole(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := h.rbacService.GetRole(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoleNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Role retrieved successfully"))
}

// @Summary List roles
// @Description List all roles with permission and user counts
// @Tags RBAC - Roles
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[RoleListItem]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles [get]
func (h *RBACHandler) ListRoles(ctx *gin.Context) {
	var req ListRolesRequest
	result, err := h.rbacService.ListRoles(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Roles retrieved successfully"))
}

// @Summary Update a role
// @Description Update an existing role
// @Tags RBAC - Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body UpdateRoleRequest true "Role update data"
// @Success 200 {object} RoleResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles/{id} [put]
func (h *RBACHandler) UpdateRole(ctx *gin.Context) {
	id := ctx.Param("id")
	var req UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.rbacService.UpdateRole(ctx, id, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Role updated successfully"))
}

// @Summary Delete a role
// @Description Delete a role by ID
// @Tags RBAC - Roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} resp.MessageResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(ctx *gin.Context) {
	id := ctx.Param("id")
	err := h.rbacService.DeleteRole(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Role deleted successfully"))
}

// ============================================================
// Permission Handlers (read-only)
// ============================================================

// @Summary List permissions
// @Description List all available permissions
// @Tags RBAC - Permissions
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[PermissionResponse]]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/permissions [get]
func (h *RBACHandler) ListPermissions(ctx *gin.Context) {
	var req ListPermissionsRequest
	result, err := h.rbacService.ListPermissions(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Permissions retrieved successfully"))
}

// ============================================================
// Role-Permission Assignment Handlers
// ============================================================

// @Summary List permissions for a role
// @Description Get all permissions assigned to a specific role
// @Tags RBAC - Role Permissions
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} resp.PaginationResponse[PermissionResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles/{id}/permissions [get]
func (h *RBACHandler) ListPermissionsForRole(ctx *gin.Context) {
	roleID := ctx.Param("id")
	result, err := h.rbacService.ListPermissionsForRole(ctx, roleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Permissions retrieved successfully"))
}

// @Summary Assign permission to role
// @Description Assign a permission to a specific role
// @Tags RBAC - Role Permissions
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param permission body AssignPermissionRequest true "Permission assignment"
// @Success 200 {object} resp.MessageResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles/{id}/permissions [post]
func (h *RBACHandler) AssignPermissionToRole(ctx *gin.Context) {
	roleID := ctx.Param("id")
	var req AssignPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	err := h.rbacService.AssignPermissionToRole(ctx, roleID, req.PermissionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Permission assigned successfully"))
}

// @Summary Remove permission from role
// @Description Remove a permission from a specific role
// @Tags RBAC - Role Permissions
// @Produce json
// @Param id path string true "Role ID"
// @Param permissionId path string true "Permission ID"
// @Success 200 {object} resp.MessageResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/roles/{id}/permissions/{permissionId} [delete]
func (h *RBACHandler) RemovePermissionFromRole(ctx *gin.Context) {
	roleID := ctx.Param("id")
	permissionID := ctx.Param("permissionId")
	err := h.rbacService.RemovePermissionFromRole(ctx, roleID, permissionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Permission removed successfully"))
}

// ============================================================
// User-Role Assignment Handlers
// ============================================================

// @Summary Assign role to user
// @Description Assign a role to a specific user
// @Tags RBAC - User Roles
// @Accept json
// @Produce json
// @Param assignment body AssignRoleToUserRequest true "User-Role assignment"
// @Success 200 {object} resp.MessageResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/user-roles [post]
func (h *RBACHandler) AssignRoleToUser(ctx *gin.Context) {
	var req AssignRoleToUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	err := h.rbacService.AssignRoleToUser(ctx, req.UserID, req.RoleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Role assigned successfully"))
}

// @Summary Remove role from user
// @Description Remove a role from a specific user
// @Tags RBAC - User Roles
// @Accept json
// @Produce json
// @Success 200 {object} resp.MessageResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/user-roles [delete]
func (h *RBACHandler) RemoveRoleFromUser(ctx *gin.Context) {
	userID := ctx.Param("userId")
	err := h.rbacService.RemoveRoleFromUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Role removed successfully"))
}

// @Summary Get role for user
// @Description Get the role assigned to a specific user
// @Tags RBAC - User Roles
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} resp.SuccessResponse[RoleResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /admin/user-roles/user/{userId} [get]
func (h *RBACHandler) GetRoleForUser(ctx *gin.Context) {
	userID := ctx.Param("userId")
	result, err := h.rbacService.GetRoleForUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Role retrieved successfully"))
}
