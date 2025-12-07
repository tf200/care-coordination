package db

import "context"

type CreateIntakeFormTxParams struct {
	IntakeForm             CreateIntakeFormParams
	RegistrationFormID     string
	RegistrationFormStatus NullRegistrationStatusEnum
}

type CreateIntakeFormTxResult struct {
	IntakeFormID string
}

func (s *Store) CreateIntakeFormTx(ctx context.Context, arg CreateIntakeFormTxParams) (CreateIntakeFormTxResult, error) {
	var result CreateIntakeFormTxResult

	err := s.ExecTx(ctx, func(q *Queries) error {
		// 1. Create the intake form
		if err := q.CreateIntakeForm(ctx, arg.IntakeForm); err != nil {
			return err
		}
		result.IntakeFormID = arg.IntakeForm.ID

		// 2. Update the registration form status to approved
		if err := q.UpdateRegistrationFormStatus(ctx, UpdateRegistrationFormStatusParams{
			ID:     arg.RegistrationFormID,
			Status: arg.RegistrationFormStatus,
		}); err != nil {
			return err
		}

		return nil
	})

	return result, err
}
