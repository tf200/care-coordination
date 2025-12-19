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

func (s *Store) CreateIntakeFormTx(
	ctx context.Context,
	arg CreateIntakeFormTxParams,
) (CreateIntakeFormTxResult, error) {
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

type UpdateIntakeFormTxParams struct {
	IntakeForm UpdateIntakeFormParams
	// If true, also update the associated client with the relevant fields
	UpdateClient bool
}

func (s *Store) UpdateIntakeFormTx(ctx context.Context, arg UpdateIntakeFormTxParams) error {
	return s.ExecTx(ctx, func(q *Queries) error {
		// 1. Update the intake form
		if err := q.UpdateIntakeForm(ctx, arg.IntakeForm); err != nil {
			return err
		}

		// 2. If requested, update the associated client with relevant fields
		if arg.UpdateClient {
			if err := q.UpdateClientByIntakeFormID(ctx, UpdateClientByIntakeFormIDParams{
				IntakeFormID:       arg.IntakeForm.ID,
				CoordinatorID:      arg.IntakeForm.CoordinatorID,
				AssignedLocationID: arg.IntakeForm.LocationID,
				FamilySituation:    arg.IntakeForm.FamilySituation,
				Limitations:        arg.IntakeForm.Limitations,
				FocusAreas:         arg.IntakeForm.FocusAreas,
				Goals:              arg.IntakeForm.Goals,
				Notes:              arg.IntakeForm.Notes,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}
