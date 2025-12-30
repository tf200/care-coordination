package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateEvaluationTxParams struct {
	Evaluation    CreateClientEvaluationParams
	ProgressLogs  []CreateGoalProgressLogParams
	IntervalWeeks int32
}

type CreateEvaluationTxResult struct {
	EvaluationID       string
	NextEvaluationDate pgtype.Date
}

func (s *Store) CreateEvaluationTx(ctx context.Context, arg CreateEvaluationTxParams) (CreateEvaluationTxResult, error) {
	var result CreateEvaluationTxResult

	err := s.ExecTx(ctx, func(q *Queries) error {
		// 1. Create evaluation bundle
		eval, err := q.CreateClientEvaluation(ctx, arg.Evaluation)
		if err != nil {
			return err
		}
		result.EvaluationID = eval.ID

		// 2. Create all progress logs
		for _, log := range arg.ProgressLogs {
			log.EvaluationID = eval.ID
			if err := q.CreateGoalProgressLog(ctx, log); err != nil {
				return err
			}
		}

		// 3. Calculate and update next evaluation date (only for submitted evaluations)
		if arg.IntervalWeeks > 0 && eval.Status == EvaluationStatusEnumSubmitted {
			nextDate := eval.EvaluationDate.Time.AddDate(0, 0, int(arg.IntervalWeeks)*7)
			result.NextEvaluationDate = pgtype.Date{Time: nextDate, Valid: true}

			if err := q.UpdateClientNextEvaluationDate(ctx, UpdateClientNextEvaluationDateParams{
				ID:                 eval.ClientID,
				NextEvaluationDate: result.NextEvaluationDate,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

type UpdateEvaluationTxParams struct {
	EvaluationID   string
	EvaluationDate pgtype.Date
	OverallNotes   *string
	ProgressLogs   []UpdateGoalProgressLogParams
}

type UpdateEvaluationTxResult struct {
	EvaluationID string
}

func (s *Store) UpdateEvaluationTx(ctx context.Context, arg UpdateEvaluationTxParams) (UpdateEvaluationTxResult, error) {
	var result UpdateEvaluationTxResult

	err := s.ExecTx(ctx, func(q *Queries) error {
		// 1. Update evaluation record
		eval, err := q.UpdateClientEvaluation(ctx, UpdateClientEvaluationParams{
			ID:             arg.EvaluationID,
			EvaluationDate: arg.EvaluationDate,
			OverallNotes:   arg.OverallNotes,
		})
		if err != nil {
			return err
		}
		result.EvaluationID = eval.ID

		// 2. Update all progress logs
		for _, log := range arg.ProgressLogs {
			if err := q.UpdateGoalProgressLog(ctx, log); err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
