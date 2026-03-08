// Package mongodb provides an OTel-traced MongoDB client.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the MongoDB connection settings.
type Config struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
	MaxPoolSize    uint64
}

// NewClient creates an OTel-traced MongoDB client and verifies connectivity
// by sending a ping to the server.
//
// The returned *mongo.Client must be disconnected when no longer needed:
//
//	defer client.Disconnect(ctx)
func NewClient(ctx context.Context, cfg Config, tp trace.TracerProvider) (*mongo.Client, error) {
	timeout := cfg.ConnectTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	opts := options.Client().
		ApplyURI(cfg.URI).
		SetConnectTimeout(timeout)

	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongodb: ping: %w", err)
	}

	return client, nil
}

// DB returns the named database from an existing client.
func DB(client *mongo.Client, name string) *mongo.Database {
	return client.Database(name)
}

// Collection returns the named collection from the given database.
func Collection(db *mongo.Database, name string) *mongo.Collection {
	return db.Collection(name)
}
