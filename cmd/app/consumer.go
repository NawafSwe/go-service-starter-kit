package app

import (
	"fmt"

	consumerapp "github.com/nawafswe/go-service-starter-kit/internal/app/transport/consumer"
	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/consumer/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/tracing"
	"github.com/nawafswe/go-service-starter-kit/internal/worker"
)

// ConsumerProcess wires up and starts the message consumer.
type ConsumerProcess struct{}

func NewConsumerProcess() ConsumerProcess { return ConsumerProcess{} }

func (c ConsumerProcess) Register(args ProcessArgs) (Process, error) {
	ctx := args.Ctx
	cfg := args.Cfg
	lgr := args.Lgr

	lgr.Info(ctx, "initializing consumer process...")

	tp, shutdown, err := tracing.Setup(ctx, cfg)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to initialize tracer")
		return nil, err
	}

	resources := bootstrap.SharedResource{Lgr: lgr}

	dbConn, err := postgres.NewConn(ctx, cfg.DB, fmt.Sprintf("%s.consumer.db", config.ServiceName), tp)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to connect to database")
		return nil, err
	}

	deps := bootstrap.Dependencies{DBConn: dbConn}

	consumer, err := consumerapp.NewConsumer(ctx, cfg, &deps, &resources)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to build consumer")
		return nil, err
	}

	consumerWorker, err := worker.NewConsumerWorker(consumer, lgr)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to create consumer worker")
		return nil, err
	}
	return withShutdown(consumerWorker, shutdown), nil
}
