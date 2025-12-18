package rbac

import (
	"care-cordination/features/middleware"
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
	// TODO: Add admin permission check: admin.Use(h.mdw.RequirePermission("admin", "manage"))

	// Roles
	roles := admin.Group("/roles")
	roles.POST("", h.CreateRole)
	roles.GET("", h.mdw.PaginationMdw(), h.ListRoles)
	roles.GET("/:id", h.GetRole)
	roles.PUT("/:id", h.UpdateRole)
	roles.DELETE("/:id", h.DeleteRole)
	roles.GET("/:id/permissions", h.ListPermissionsForRole)
	roles.POST("/:id/permissions", h.AssignPermissionToRole)
	roles.DELETE("/:id/permissions/:permissionId", h.RemovePermissionFromRole)

	// Permissions (read-only)
	permissions := admin.Group("/permissions")
	permissions.GET("", h.mdw.PaginationMdw(), h.ListPermissions)

	// User-Role assignments
	userRoles := admin.Group("/user-roles")
	userRoles.POST("", h.AssignRoleToUser)
	userRoles.DELETE("/user/:userId", h.RemoveRoleFromUser)
	userRoles.GET("/user/:userId", h.GetRoleForUser)
}

// ============================================================
// Role Handlers
// ============================================================

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
	ctx.JSON(http.StatusCreated, result)
}

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
	ctx.JSON(http.StatusOK, result)
}

func (h *RBACHandler) ListRoles(ctx *gin.Context) {
	var req ListRolesRequest
	result, err := h.rbacService.ListRoles(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

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
	ctx.JSON(http.StatusOK, result)
}

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

func (h *RBACHandler) ListPermissions(ctx *gin.Context) {
	var req ListPermissionsRequest
	result, err := h.rbacService.ListPermissions(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// ============================================================
// Role-Permission Assignment Handlers
// ============================================================

func (h *RBACHandler) ListPermissionsForRole(ctx *gin.Context) {
	roleID := ctx.Param("id")
	result, err := h.rbacService.ListPermissionsForRole(ctx, roleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

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

func (h *RBACHandler) RemoveRoleFromUser(ctx *gin.Context) {
	userID := ctx.Param("userId")
	err := h.rbacService.RemoveRoleFromUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Role removed successfully"))
}

func (h *RBACHandler) GetRoleForUser(ctx *gin.Context) {
	userID := ctx.Param("userId")
	result, err := h.rbacService.GetRoleForUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}
