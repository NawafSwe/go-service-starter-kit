// Package tracing sets up OpenTelemetry distributed tracing and log export.
package tracing

import (
	"context"
	"fmt"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelloggrpc "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	oteltracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	otelsdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ShutdownFunc must be called when the application exits to flush and close
// the exporter connections cleanly.
type ShutdownFunc func(context.Context) error

// Setup initialises the global OTel trace provider and log provider, then
// returns the provider and a shutdown function that flushes all pending
// spans/logs and closes the underlying gRPC connection.
//
// When tracing is disabled in config the global no-op provider is returned
// and the shutdown function is a safe no-op — callers do not need an
// if-statement around Setup.
func Setup(ctx context.Context, cfg config.Config) (trace.TracerProvider, ShutdownFunc, error) {
	if !cfg.General.Tracing.Enabled {
		return otel.GetTracerProvider(), func(context.Context) error { return nil }, nil
	}

	grpcConn, err := grpc.NewClient(
		cfg.General.Tracing.ReceiverEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("tracing: dial collector: %w", err)
	}

	// ---- trace exporter ----
	traceExporter, err := oteltracegrpc.New(ctx, oteltracegrpc.WithGRPCConn(grpcConn))
	if err != nil {
		_ = grpcConn.Close()
		return nil, nil, fmt.Errorf("tracing: create trace exporter: %w", err)
	}

	// ---- log exporter ----
	logExporter, err := otelloggrpc.New(ctx, otelloggrpc.WithGRPCConn(grpcConn))
	if err != nil {
		_ = grpcConn.Close()
		return nil, nil, fmt.Errorf("tracing: create log exporter: %w", err)
	}

	// ---- resource ----
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.General.ServiceName),
			semconv.ServiceVersionKey.String(cfg.General.AppVersion),
			semconv.DeploymentEnvironmentName(cfg.General.AppEnvironment),
		),
	)
	if err != nil {
		_ = grpcConn.Close()
		return nil, nil, fmt.Errorf("tracing: create resource: %w", err)
	}

	// ---- trace provider ----
	tp := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(traceExporter),
		otelsdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// ---- log provider ----
	lp := otelsdklog.NewLoggerProvider(
		otelsdklog.WithProcessor(otelsdklog.NewBatchProcessor(logExporter)),
	)
	global.SetLoggerProvider(lp)

	shutdown := func(ctx context.Context) error {
		var errs []error
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("trace provider shutdown: %w", err))
		}
		if err := lp.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("log provider shutdown: %w", err))
		}
		if err := grpcConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("grpc conn close: %w", err))
		}
		if len(errs) > 0 {
			return fmt.Errorf("tracing shutdown errors: %v", errs)
		}
		return nil
	}

	return tp, shutdown, nil
}

// StartSpan is a convenience helper that starts a named span on the global tracer.
// The caller is responsible for calling span.End().
//
//	ctx, span := tracing.StartSpan(ctx, "my-component", "do-thing",
//	    oteltracer.WithSpanKind(oteltracer.SpanKindServer))
//	defer span.End()
func StartSpan(ctx context.Context, tracerName, spanName string, attrs []attribute.KeyValue, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, spanName, opts...)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	return ctx, span
}

// FailSpan records err on span and sets its status to Error.
// Returns err unchanged so callers can write:
//
//	return tracing.FailSpan(span, err)
func FailSpan(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}
