package rbac

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInternal           = errors.New("internal server error")
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrRoleAlreadyExists  = errors.New("role already exists")
)
