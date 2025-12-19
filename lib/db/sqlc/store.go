package db

import (
	"care-cordination/lib/util"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*Queries
	ConnPool *pgxpool.Pool
}

func NewStore(connPool *pgxpool.Pool) *Store {
	return &Store{
		Queries:  New(connPool),
		ConnPool: connPool,
	}
}

func (store *Store) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	// 1. Start a transaction
	tx, err := store.ConnPool.Begin(ctx)
	if err != nil {
		return err
	}

	// Set RLS context if user is authenticated
	userID := util.GetUserID(ctx)
	if userID != "" {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to set rls context: %w", err)
		}
	}

	// 2. Create a new Queries instance using that specific transaction
	q := New(tx)

	// 3. Run the custom business logic (the callback function)
	err = fn(q)

	// 4. Handle Rollback or Commit
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
