// Package mongodb provides a MongoDB implementation of the database.Database interface.
// It is intended to use go.mongodb.org/mongo-driver as the underlying driver.
package mongodb

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/database"
)

// Adapter implements the database.Database interface using MongoDB.
//
// Intended dependency: go.mongodb.org/mongo-driver
type Adapter struct {
	// client *mongo.Client // TODO: uncomment when mongo-driver dependency is added
}

// New creates a new MongoDB Adapter instance.
func New() *Adapter {
	return &Adapter{}
}

// Connect establishes a connection to the MongoDB cluster.
func (a *Adapter) Connect(ctx context.Context, cfg database.Config) error {
	// TODO: implement using mongo-driver
	return nil
}

// Close disconnects from the MongoDB cluster.
func (a *Adapter) Close() error {
	// TODO: implement using mongo-driver
	return nil
}

// Ping verifies the MongoDB connection is alive.
func (a *Adapter) Ping(ctx context.Context) error {
	// TODO: implement using mongo-driver
	return nil
}

// Transaction returns a new Transaction backed by a MongoDB session.
func (a *Adapter) Transaction() database.Transaction {
	// TODO: implement using mongo-driver
	return nil
}
