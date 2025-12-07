package db

import "context"

type UpdateRegistrationFormTxParams struct {
	RegistrationForm UpdateRegistrationFormParams
	// If true, also update the associated client with the relevant fields
	UpdateClient bool
}

func (s *Store) UpdateRegistrationFormTx(ctx context.Context, arg UpdateRegistrationFormTxParams) error {
	return s.ExecTx(ctx, func(q *Queries) error {
		// 1. Update the registration form
		if err := q.UpdateRegistrationForm(ctx, arg.RegistrationForm); err != nil {
			return err
		}

		// 2. If requested, update the associated client with relevant fields
		if arg.UpdateClient {
			if err := q.UpdateClientByRegistrationFormID(ctx, UpdateClientByRegistrationFormIDParams{
				RegistrationFormID: arg.RegistrationForm.ID,
				FirstName:          arg.RegistrationForm.FirstName,
				LastName:           arg.RegistrationForm.LastName,
				Bsn:                arg.RegistrationForm.Bsn,
				DateOfBirth:        arg.RegistrationForm.DateOfBirth,
				Gender:             arg.RegistrationForm.Gender,
				CareType:           arg.RegistrationForm.CareType,
				ReferringOrgID:     arg.RegistrationForm.RefferingOrgID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}
