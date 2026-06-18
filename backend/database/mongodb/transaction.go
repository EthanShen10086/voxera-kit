package mongodb

import (
	"context"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/database"
	"go.mongodb.org/mongo-driver/mongo"
)

// Transaction implements database.Transaction using MongoDB sessions.
type Transaction struct {
	client  *mongo.Client
	session mongo.Session
	started bool
}

// Begin starts a new MongoDB transaction session.
func (t *Transaction) Begin(ctx context.Context) (database.Transaction, error) {
	if t.client == nil {
		return nil, fmt.Errorf("mongodb: not connected")
	}
	if t.session != nil {
		return t, nil
	}

	session, err := t.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("mongodb: start session: %w", err)
	}
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("mongodb: start transaction: %w", err)
	}

	return &Transaction{
		client:  t.client,
		session: session,
		started: true,
	}, nil
}

// Commit finalizes the MongoDB transaction.
func (t *Transaction) Commit() error {
	if !t.started || t.session == nil {
		return fmt.Errorf("mongodb: transaction not started")
	}

	ctx := context.Background()
	err := t.session.CommitTransaction(ctx)
	t.session.EndSession(ctx)
	t.session = nil
	t.started = false
	return err
}

// Rollback aborts the MongoDB transaction.
func (t *Transaction) Rollback() error {
	if !t.started || t.session == nil {
		return fmt.Errorf("mongodb: transaction not started")
	}

	ctx := context.Background()
	err := t.session.AbortTransaction(ctx)
	t.session.EndSession(ctx)
	t.session = nil
	t.started = false
	return err
}
