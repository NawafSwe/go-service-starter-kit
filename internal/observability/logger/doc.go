// Package logger provides a structured, context-aware logger backed by zerolog.
//
// # Creating a logger
//
//	lgr := logger.NewLogger(logger.InfoLevel, "my-service", "1.0.0", "production")
//
// Builder methods return a new logger value — the original is unchanged,
// making it safe to share across goroutines:
//
//	lgr = lgr.WithHost(hostname).WithFields(map[string]any{"component": "worker"})
//
// # Context fields
//
// Inject request-scoped fields once and they appear on every log line that
// uses the same context — no need to thread fields through function signatures:
//
//	ctx = logger.ContextWithFields(ctx, map[string]any{"request_id": id})
//	lgr.Info(ctx, "processing")   // → includes request_id automatically
//
// # Interface
//
// Depend on the Logger interface, not the concrete ZerologLogger, so
// you can swap implementations or use mocks in tests:
//
//	type Handler struct { lgr logger.Logger }
package logger
