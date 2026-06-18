// Package mongodb provides a MongoDB implementation of the database.Database interface.
// It uses go.mongodb.org/mongo-driver as the underlying driver.
package mongodb

import (
	"context"
	"fmt"

	"github.com/EthanShen10086/voxera-kit/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Adapter implements the database.Database interface using MongoDB.
type Adapter struct {
	client *mongo.Client
}

// New creates a new MongoDB Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// Connect establishes a connection to the MongoDB cluster.
func (a *Adapter) Connect(ctx context.Context, cfg database.Config) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri(cfg)))
	if err != nil {
		return fmt.Errorf("mongodb: connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return fmt.Errorf("mongodb: ping: %w", err)
	}

	if a.client != nil {
		_ = a.client.Disconnect(ctx)
	}
	a.client = client
	return nil
}

// Close disconnects from the MongoDB cluster.
func (a *Adapter) Close() error {
	if a.client == nil {
		return nil
	}
	err := a.client.Disconnect(context.Background())
	a.client = nil
	return err
}

// Ping verifies the MongoDB connection is alive.
func (a *Adapter) Ping(ctx context.Context) error {
	if a.client == nil {
		return fmt.Errorf("mongodb: not connected")
	}
	return a.client.Ping(ctx, nil)
}

// Transaction returns a new Transaction backed by a MongoDB session.
func (a *Adapter) Transaction() database.Transaction {
	return &Transaction{client: a.client}
}

func uri(cfg database.Config) string {
	port := cfg.Port
	if port == 0 {
		port = 27017
	}

	creds := ""
	if cfg.User != "" {
		creds = fmt.Sprintf("%s:%s@", cfg.User, cfg.Password)
	}

	db := cfg.Database
	if db == "" {
		db = "admin"
	}

	return fmt.Sprintf("mongodb://%s%s:%d/%s", creds, cfg.Host, port, db)
}
