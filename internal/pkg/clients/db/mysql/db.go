// Package mysql provides an OTel-traced MySQL connection pool backed by sqlx.
package mysql

import (
	"context"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

const driver = "mysql"

// NewConn creates an OTel-traced MySQL connection pool.
// DSN format: user:password@tcp(host:port)/dbname?parseTime=true
// The connection is pinged during setup — an error is returned if the database
// is unreachable.
func NewConn(ctx context.Context, cfg config.DB, connectionName string, tp trace.TracerProvider) (*sqlx.DB, error) {
	tracedDB, err := otelsql.Open(driver, cfg.DSN, otelsql.WithTracerProvider(tp))
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}
	_, err = otelsql.RegisterDBStatsMetrics(tracedDB, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
		semconv.ServiceNameKey.String(connectionName),
	))
	if err != nil {
		return nil, fmt.Errorf("mysql: register stats metrics: %w", err)
	}

	db := sqlx.NewDb(tracedDB, driver)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.MaxConnectionsLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConnections)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}
	return db, nil
}
