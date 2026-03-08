package bootstrap

import (
	"github.com/nawafswe/go-service-starter-kit/internal/app/repositories/example"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
)

// InitializeRepositories wires up and returns the shared repository instances.
func InitializeRepositories(_ config.Config, deps *Dependencies, _ *SharedResource) (*SharedRepositories, error) {
	return &SharedRepositories{
		ExampleRepository: example.New(deps.DBConn),
	}, nil
}
