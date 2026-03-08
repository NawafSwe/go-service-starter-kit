package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nawafswe/go-service-starter-kit/cmd/app"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	"go.opentelemetry.io/otel"
	otelmeter "go.opentelemetry.io/otel/metric"
)

const (
	DefaultEnv        = "development"
	DefaultVer        = "1.0.0"
	DefaultConfigPath = "./config.yaml"
	DefaultEnvPath    = "./.env"
)

var (
	configPath string
	dotEnvPath string

	usageFunc = func() {
		fmt.Printf("Usage: %s [flags] [process]\n\n", os.Args[0])
		fmt.Printf("Processes:\n")
		fmt.Printf("  http\t\t\tStart HTTP server\n")
		fmt.Printf("  grpc\t\t\tStart gRPC server\n")
		fmt.Printf("  consumer\t\tStart message consumer\n")
		fmt.Printf("  <job-name>\t\tRun a one-time job\n")
		fmt.Printf("\nFlags:\n")
		flag.PrintDefaults()
	}
)

func main() {
	flag.Usage = usageFunc
	flag.StringVar(&configPath, "config", DefaultConfigPath, "Path to config YAML file")
	flag.StringVar(&dotEnvPath, "dotEnv", DefaultEnvPath, "Path to .env file")
	flag.Parse()

	ctx := context.Background()

	nonFlagArgs := flag.Args()
	if len(nonFlagArgs) < 1 {
		flag.Usage()
		return
	}
	processName := nonFlagArgs[0]

	env := envOrDefault("APP_ENVIRONMENT", DefaultEnv)
	version := envOrDefault("APP_VERSION", DefaultVer)
	lgr := logger.NewLogger(logger.InfoLevel, config.ServiceName, version, env)
	lgr.Info(ctx, fmt.Sprintf("starting %s...", config.ServiceName))

	configAbsPath, _ := filepath.Abs(configPath)
	dotEnvAbsPath, _ := filepath.Abs(dotEnvPath)
	cfg, err := config.Load(configAbsPath, dotEnvAbsPath)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to load config")
		return
	}
	lgr.Info(ctx, "config loaded successfully")

	args := app.ProcessArgs{
		Ctx: ctx,
		Lgr: lgr,
		Cfg: cfg,
	}

	if cfg.Metrics.Enabled {
		args.MeterProvider = meterProvider()
		args.Meter = meter(config.ServiceName)
	}

	// One-time jobs run and exit.
	if job, ok := app.JobsMap[processName]; ok {
		lgr.Info(ctx, fmt.Sprintf("running job %s", processName))
		if err := job.Schedule(args); err != nil {
			lgr.Error(ctx, err, "[FATAL] job failed")
		}
		lgr.Info(ctx, "job finished, exiting")
		return
	}

	// Long-running processes.
	process, ok := app.RegistryProcessesMap[processName]
	if !ok {
		lgr.Error(ctx, fmt.Errorf("unsupported process"), fmt.Sprintf("[FATAL] process %q not registered", processName))
		return
	}

	runnableProcess, err := process.Register(args)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to initialize process")
		return
	}
	if err = runnableProcess.Run(ctx); err != nil {
		lgr.Error(ctx, err, "process exited with error")
	}
}

func meterProvider() otelmeter.MeterProvider {
	mp := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	return mp
}

func meter(name string, opts ...otelmeter.MeterOption) otelmeter.Meter {
	return otel.Meter(name, opts...)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
