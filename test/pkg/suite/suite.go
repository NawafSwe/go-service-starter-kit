package suite

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/app/repositories/example"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	"github.com/nawafswe/go-service-starter-kit/test/pkg/testdb"
)

// Suite holds the shared test dependencies for an integration test.
// Each test gets its own isolated database so tests can run in parallel.
type Suite struct {
	Cfg   config.Config
	SqlDB *sqlx.DB
	Lgr   logger.Logger

	Repos Repositories
}

// Repositories holds initialised repository instances for direct test seeding
// and assertion.
type Repositories struct {
	ExampleRepository example.Repository
}

// SetupTestSuite creates an isolated test environment.
// It provisions a unique database via testdb.CreateTestDB, runs migrations,
// and initialises all repositories. Cleanup is automatic via t.Cleanup.
func SetupTestSuite(t *testing.T, opts ...Option) *Suite {
	t.Helper()

	cfg, err := config.Load("../../../config.yaml", "../../../test/.env.integration")
	if err != nil {
		t.Fatalf("suite: load config: %v", err)
	}

	lgr := logger.NewLogger(logger.ErrorLevel, cfg.General.ServiceName, cfg.General.AppVersion, cfg.General.AppEnvironment)

	dbConf, err := testdb.ParseDSN(cfg.DB.DSN)
	if err != nil {
		t.Fatalf("suite: parse dsn: %v", err)
	}

	db, err := testdb.CreateTestDB(context.Background(), t, dbConf)
	if err != nil {
		t.Fatalf("suite: create test db: %v", err)
	}

	s := &Suite{
		Cfg:   cfg,
		SqlDB: db,
		Lgr:   lgr,
		Repos: Repositories{
			ExampleRepository: example.New(db),
		},
	}

	for _, opt := range opts {
		opt(t, s)
	}

	return s
}
