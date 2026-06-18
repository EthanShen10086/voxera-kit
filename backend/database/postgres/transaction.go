package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transaction implements database.Transaction using pgx transactions and savepoints.
type Transaction struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

// Begin starts a new transaction or nested savepoint.
func (t *Transaction) Begin(ctx context.Context) (database.Transaction, error) {
	if t.pool == nil {
		return nil, fmt.Errorf("postgres: not connected")
	}

	if t.tx == nil {
		tx, err := t.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("postgres: begin transaction: %w", err)
		}
		return &Transaction{pool: t.pool, tx: tx}, nil
	}

	tx, err := t.tx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: begin savepoint: %w", err)
	}
	return &Transaction{pool: t.pool, tx: tx}, nil
}

// Commit finalizes the transaction.
func (t *Transaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("postgres: transaction not started")
	}
	return t.tx.Commit(context.Background())
}

// Rollback aborts the transaction.
func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return fmt.Errorf("postgres: transaction not started")
	}
	err := t.tx.Rollback(context.Background())
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}
