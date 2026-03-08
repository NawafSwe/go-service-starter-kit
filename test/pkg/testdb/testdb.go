// Package testdb provides helpers for creating isolated PostgreSQL databases
// in integration tests. Each test gets its own database with migrations applied,
// and the database is automatically dropped when the test completes.
package testdb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DBConfig holds the connection parameters for a PostgreSQL database.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DSN returns a PostgreSQL connection string.
func (dc DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dc.User,
		dc.Password,
		net.JoinHostPort(dc.Host, dc.Port),
		dc.DBName,
	)
}

// ParseDSN parses a PostgreSQL DSN into a DBConfig.
func ParseDSN(dsn string) (DBConfig, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return DBConfig{}, fmt.Errorf("parse dsn: %w", err)
	}

	user := parsed.User.Username()
	password, _ := parsed.User.Password()

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		// No port specified — treat the whole thing as the host.
		host = parsed.Host
		port = "5432"
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")

	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
	}, nil
}

// createConfig holds options for CreateTestDB.
type createConfig struct {
	migrationsPath string
	skipMigrations bool
}

// Option customises the behaviour of CreateTestDB.
type Option func(*createConfig)

// WithMigrationsPath overrides the default migrations directory.
func WithMigrationsPath(path string) Option {
	return func(c *createConfig) {
		c.migrationsPath = path
	}
}

// WithSkipMigrations skips running migrations — useful for testing missing-table
// error paths.
func WithSkipMigrations() Option {
	return func(c *createConfig) {
		c.skipMigrations = true
	}
}

// CreateTestDB provisions an isolated PostgreSQL database for a single test.
//
// It creates a uniquely-named database, optionally applies migrations, and
// returns a connection to it. The database is automatically dropped when the
// test completes.
//
//nolint:thelper // t is used for cleanup, not as a test entry point.
func CreateTestDB(ctx context.Context, t testing.TB, conf DBConfig, opts ...Option) (*sqlx.DB, error) {
	cfg := buildConfig(opts...)

	adminConn, err := sqlx.ConnectContext(ctx, "postgres", conf.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to admin db: %w", err)
	}

	dbName := "test_" + strings.ReplaceAll(uuid.New().String(), "-", "_")
	if _, err := adminConn.ExecContext(ctx, "CREATE DATABASE "+dbName); err != nil {
		adminConn.Close()
		return nil, fmt.Errorf("create database %s: %w", dbName, err)
	}

	testConf := conf
	testConf.DBName = dbName

	if !cfg.skipMigrations {
		if err := applyMigrations(testConf, cfg.migrationsPath); err != nil {
			// Best-effort cleanup on migration failure.
			dropDB(context.Background(), adminConn, dbName)
			adminConn.Close()
			return nil, fmt.Errorf("apply migrations: %w", err)
		}
	}

	testConn, err := sqlx.ConnectContext(ctx, "postgres", testConf.DSN())
	if err != nil {
		dropDB(context.Background(), adminConn, dbName)
		adminConn.Close()
		return nil, fmt.Errorf("connect to test db %s: %w", dbName, err)
	}

	t.Cleanup(func() {
		testConn.Close()
		cleanupCtx := context.Background()
		if err := dropDB(cleanupCtx, adminConn, dbName); err != nil {
			t.Errorf("testdb: drop %s: %v", dbName, err)
		}
		adminConn.Close()
	})

	return testConn, nil
}

func buildConfig(opts ...Option) createConfig {
	cfg := createConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.migrationsPath == "" {
		path, err := findMigrationsDir()
		if err == nil {
			cfg.migrationsPath = path
		}
	}
	return cfg
}

// applyMigrations runs all up migrations against the given database.
func applyMigrations(conf DBConfig, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, conf.DSN())
	if err != nil {
		return fmt.Errorf("open migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// dropDB terminates active connections and drops the database.
func dropDB(ctx context.Context, conn *sqlx.DB, name string) error {
	// Terminate active sessions so DROP doesn't block.
	_, _ = conn.ExecContext(ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", name)

	if _, err := conn.ExecContext(ctx, "DROP DATABASE IF EXISTS "+name); err != nil {
		return fmt.Errorf("drop database %s: %w", name, err)
	}
	return nil
}

// findMigrationsDir walks upward from the working directory looking for
// db/migrations (the project convention).
func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	const maxDepth = 7
	for range maxDepth {
		candidate := filepath.Join(dir, "db", "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("db/migrations directory not found")
}
