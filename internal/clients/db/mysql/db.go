// Package mysql provides a MySQL connection pool backed by sqlx.
// When a TracerProvider is supplied via WithTracerProvider, queries are
// automatically traced with OTel.
package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

const driver = "mysql"

// Option configures the MySQL connection.
type Option func(*connConfig)

type connConfig struct {
	tp trace.TracerProvider
}

// WithTracerProvider enables OTel tracing on all SQL operations.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *connConfig) { c.tp = tp }
}

// NewConn creates a MySQL connection pool. By default, no tracing is applied.
// Use WithTracerProvider to enable OTel-traced queries.
// DSN format: user:password@tcp(host:port)/dbname?parseTime=true
// The connection is pinged during setup — an error is returned if the database
// is unreachable.
func NewConn(ctx context.Context, cfg config.DB, connectionName string, opts ...Option) (*sqlx.DB, error) {
	var cc connConfig
	for _, opt := range opts {
		opt(&cc)
	}

	var rawDB *sql.DB
	if cc.tp != nil {
		traced, err := otelsql.Open(driver, cfg.DSN, otelsql.WithTracerProvider(cc.tp))
		if err != nil {
			return nil, fmt.Errorf("mysql: open (traced): %w", err)
		}
		_, err = otelsql.RegisterDBStatsMetrics(traced, otelsql.WithAttributes(
			semconv.DBSystemMySQL,
			semconv.ServiceNameKey.String(connectionName),
		))
		if err != nil {
			return nil, fmt.Errorf("mysql: register stats metrics: %w", err)
		}
		rawDB = traced
	} else {
		plain, err := sql.Open(driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("mysql: open: %w", err)
		}
		rawDB = plain
	}

	db := sqlx.NewDb(rawDB, driver)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.MaxConnectionsLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConnections)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}
	return db, nil
}
