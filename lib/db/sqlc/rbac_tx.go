package db

import "context"

type CreateRoleWithPermissionsTxParams struct {
	Role          CreateRoleParams
	PermissionIDs []string
}

type CreateRoleWithPermissionsTxResult struct {
	Role Role
}

func (s *Store) CreateRoleWithPermissionsTx(
	ctx context.Context,
	arg CreateRoleWithPermissionsTxParams,
) (CreateRoleWithPermissionsTxResult, error) {
	var result CreateRoleWithPermissionsTxResult

	err := s.ExecTx(ctx, func(q *Queries) error {
		// 1. Create the role
		role, err := q.CreateRole(ctx, arg.Role)
		if err != nil {
			return err
		}
		result.Role = role

		// 2. If there are permissions, batch assign them to the role
		if len(arg.PermissionIDs) > 0 {
			if err := q.BatchAssignPermissionsToRole(ctx, BatchAssignPermissionsToRoleParams{
				RoleID:        role.ID,
				PermissionIds: arg.PermissionIDs,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

type UpdateRoleWithPermissionsTxParams struct {
	Role          UpdateRoleParams
	PermissionIDs []string
}

type UpdateRoleWithPermissionsTxResult struct {
	Role Role
}

func (s *Store) UpdateRoleWithPermissionsTx(
	ctx context.Context,
	arg UpdateRoleWithPermissionsTxParams,
) (UpdateRoleWithPermissionsTxResult, error) {
	var result UpdateRoleWithPermissionsTxResult

	err := s.ExecTx(ctx, func(q *Queries) error {
		// 1. Update the role
		role, err := q.UpdateRole(ctx, arg.Role)
		if err != nil {
			return err
		}
		result.Role = role

		// 2. Delete all existing permissions for this role
		if err := q.DeleteAllPermissionsFromRole(ctx, role.ID); err != nil {
			return err
		}

		// 3. If there are permissions, batch assign them to the role
		if len(arg.PermissionIDs) > 0 {
			if err := q.BatchAssignPermissionsToRole(ctx, BatchAssignPermissionsToRoleParams{
				RoleID:        role.ID,
				PermissionIds: arg.PermissionIDs,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
