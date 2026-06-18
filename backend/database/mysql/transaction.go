package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/database"
)

// Transaction implements database.Transaction using sql.Tx.
type Transaction struct {
	db *sql.DB
	tx *sql.Tx
}

// Begin starts a new transaction or nested savepoint.
func (t *Transaction) Begin(ctx context.Context) (database.Transaction, error) {
	if t.db == nil {
		return nil, fmt.Errorf("mysql: not connected")
	}

	if t.tx == nil {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("mysql: begin transaction: %w", err)
		}
		return &Transaction{db: t.db, tx: tx}, nil
	}

	if _, err := t.tx.ExecContext(ctx, "SAVEPOINT sp"); err != nil {
		return nil, fmt.Errorf("mysql: create savepoint: %w", err)
	}
	return &Transaction{db: t.db, tx: t.tx}, nil
}

// Commit finalizes the transaction.
func (t *Transaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("mysql: transaction not started")
	}
	return t.tx.Commit()
}

// Rollback aborts the transaction.
func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return fmt.Errorf("mysql: transaction not started")
	}
	return t.tx.Rollback()
}
