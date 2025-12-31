package db

import (
	"context"
)

// StoreInterface defines the contract for database operations.
// It embeds Querier for all query methods and adds transaction support.
//
//go:generate mockgen -destination=mocks/mock_store.go -package=mocks care-cordination/lib/db/sqlc StoreInterface
type StoreInterface interface {
	Querier

	// Transaction methods
	ExecTx(ctx context.Context, fn func(*Queries) error) error

	// Evaluation transaction
	CreateEvaluationTx(ctx context.Context, params CreateEvaluationTxParams) (CreateEvaluationTxResult, error)
	UpdateEvaluationTx(ctx context.Context, params UpdateEvaluationTxParams) (UpdateEvaluationTxResult, error)

	// Client transaction
	MoveClientToWaitingListTx(ctx context.Context, arg MoveClientToWaitingListTxParams) (MoveClientToWaitingListTxResult, error)

	// Employee transaction
	CreateEmployeeTx(ctx context.Context, arg CreateEmployeeTxParams) error
}

// Ensure Store implements StoreInterface
var _ StoreInterface = (*Store)(nil)
