package bootstrap

import (
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
)

// InitializeClients initialises all external client connections and returns them
// wrapped in a Dependencies struct.
func InitializeClients(_ config.Config, db *sqlx.DB, _ *SharedResource) (Dependencies, error) {
	return Dependencies{
		DBConn: db,
	}, nil
}
