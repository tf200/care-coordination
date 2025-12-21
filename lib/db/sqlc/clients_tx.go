package db

import "context"

type MoveClientToWaitingListTxParams struct {
	Client              CreateClientParams
	IntakeFormID        string
	IntakeFormNewStatus IntakeStatusEnum
}

type MoveClientToWaitingListTxResult struct {
	ClientID string
}

func (s *Store) MoveClientToWaitingListTx(
	ctx context.Context,
	arg MoveClientToWaitingListTxParams,
) (MoveClientToWaitingListTxResult, error) {
	var result MoveClientToWaitingListTxResult

	err := s.ExecTx(ctx, func(q *Queries) error {
		// 1. Create the client
		client, err := q.CreateClient(ctx, arg.Client)
		if err != nil {
			return err
		}
		result.ClientID = client.ID

		// 2. Update the intake form status
		if err := q.UpdateIntakeFormStatus(ctx, UpdateIntakeFormStatusParams{
			ID:     arg.IntakeFormID,
			Status: arg.IntakeFormNewStatus,
		}); err != nil {
			return err
		}

		// 3. Link goals to the new client
		if err := q.LinkGoalsToClient(ctx, LinkGoalsToClientParams{
			ClientID:     &client.ID,
			IntakeFormID: arg.IntakeFormID,
		}); err != nil {
			return err
		}

		return nil
	})

	return result, err
}
