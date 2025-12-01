package db

import "context"

type CreateEmployeeTxParams struct {
	Emp  CreateEmployeeParams
	User CreateUserParams
}

func (s *Store) CreateEmployeeTx(ctx context.Context, arg CreateEmployeeTxParams) error {
	err := s.ExecTx(ctx, func(q *Queries) error {
		userID, err := q.CreateUser(ctx, arg.User)
		if err != nil {
			return err
		}
		arg.Emp.UserID = userID
		if err := q.CreateEmployee(ctx, arg.Emp); err != nil {
			return err
		}
		return nil
	})
	return err
}
