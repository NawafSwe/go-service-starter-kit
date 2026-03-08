// Package postgres provides an OTel-traced PostgreSQL connection pool backed by sqlx.
package postgres

import (
	"context"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

const driver = "postgres"

// NewConn creates an OTel-traced PostgreSQL connection pool.
// The connection is pinged during setup — an error is returned if the database
// is unreachable.
func NewConn(ctx context.Context, cfg config.DB, connectionName string, tp trace.TracerProvider) (*sqlx.DB, error) {
	tracedDB, err := otelsql.Open(driver, cfg.DSN, otelsql.WithTracerProvider(tp))
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	_, err = otelsql.RegisterDBStatsMetrics(tracedDB, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.ServiceNameKey.String(connectionName),
	))
	if err != nil {
		return nil, fmt.Errorf("postgres: register stats metrics: %w", err)
	}

	db := sqlx.NewDb(tracedDB, driver)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.MaxConnectionsLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConnections)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}
