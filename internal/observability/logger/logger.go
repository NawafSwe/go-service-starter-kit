package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate go tool mockgen -source=${GOFILE} -destination=mock/${GOFILE} -package=mock

// Logger is the primary logging interface.
// All application code should depend on this interface, never on ZerologLogger directly.
type Logger interface {
	Debug(ctx context.Context, msg string)
	DebugW(ctx context.Context, msg string, fields map[string]any)
	Info(ctx context.Context, msg string)
	InfoW(ctx context.Context, msg string, fields map[string]any)
	Warn(ctx context.Context, msg string)
	WarnW(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, err error, msg string)
	ErrorW(ctx context.Context, err error, msg string, fields map[string]any)
}

// Level enumerates available log levels.
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"

	fieldMinLevel    = "minimum_log_level"
	fieldApplication = "application"
	fieldHost        = "host"
	fieldVersion     = "version"
	fieldEnvironment = "environment"

	// callerSkip is the number of stack frames to skip so the caller field
	// points at the call site, not at this package.
	callerSkip = 2
)

// ZerologLogger is the zerolog-backed implementation of Logger.
// Construct one with NewLogger; pass it around as the Logger interface.
// Builder methods (WithAttributes, WithHost, …) return a new value — the
// original is unchanged, which makes it safe to share across goroutines.
type ZerologLogger struct {
	root       zerolog.Logger
	callerSkip int
}

//nolint:gochecknoinits
func init() {
	// Use nanosecond precision so high-frequency logs remain sortable.
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

// NewLogger creates a ZerologLogger with the given minimum level and
// static service metadata attached to every log line.
func NewLogger(minLevel Level, appName, version, env string) ZerologLogger {
	root := log.Logger.With().
		Str(fieldApplication, appName).
		Str(fieldVersion, version).
		Str(fieldEnvironment, env).
		Str(fieldMinLevel, string(minLevel)).
		Logger().
		Level(toZerologLevel(minLevel))
	return ZerologLogger{root: root, callerSkip: callerSkip}
}

// WithFields returns a new logger with extra static fields attached.
func (l ZerologLogger) WithFields(fields map[string]any) ZerologLogger {
	l.root = l.root.With().Fields(fields).Logger()
	return l
}

// WithHost returns a new logger with the host field set.
func (l ZerologLogger) WithHost(host string) ZerologLogger {
	l.root = l.root.With().Str(fieldHost, host).Logger()
	return l
}

// WithLevel returns a new logger filtered to the given minimum level.
func (l ZerologLogger) WithLevel(minLevel Level) ZerologLogger {
	l.root = l.root.With().
		Str(fieldMinLevel, string(minLevel)).
		Logger().
		Level(toZerologLevel(minLevel))
	return l
}

// WithOutput returns a new logger writing to w instead of stderr.
func (l ZerologLogger) WithOutput(w io.Writer) ZerologLogger {
	l.root = l.root.Output(w)
	return l
}

// WithPrettyOutput returns a new logger with human-readable console output.
// Use this in local development; JSON output is better in production.
func (l ZerologLogger) WithPrettyOutput() ZerologLogger {
	return l.WithOutput(zerolog.ConsoleWriter{
		Out:          os.Stderr,
		PartsExclude: []string{zerolog.CallerFieldName},
		FieldsExclude: []string{
			fieldMinLevel, fieldApplication,
			fieldHost, fieldVersion, fieldEnvironment,
		},
		TimeFormat: "Jan _2 15:04:05.000",
	})
}

// WithCallerSkip returns a new logger with an adjusted caller frame offset.
// Use this when wrapping ZerologLogger in another logger type.
func (l ZerologLogger) WithCallerSkip(skip int) ZerologLogger {
	l.callerSkip = skip
	return l
}

// ---- Logger interface implementation ----

func (l ZerologLogger) Debug(ctx context.Context, msg string) {
	l.newEvent(ctx, zerolog.DebugLevel, nil).Msg(msg)
}

func (l ZerologLogger) DebugW(ctx context.Context, msg string, fields map[string]any) {
	l.newEvent(ctx, zerolog.DebugLevel, fields).Msg(msg)
}

func (l ZerologLogger) Info(ctx context.Context, msg string) {
	l.newEvent(ctx, zerolog.InfoLevel, nil).Msg(msg)
}

func (l ZerologLogger) InfoW(ctx context.Context, msg string, fields map[string]any) {
	l.newEvent(ctx, zerolog.InfoLevel, fields).Msg(msg)
}

func (l ZerologLogger) Warn(ctx context.Context, msg string) {
	l.newEvent(ctx, zerolog.WarnLevel, nil).Msg(msg)
}

func (l ZerologLogger) WarnW(ctx context.Context, msg string, fields map[string]any) {
	l.newEvent(ctx, zerolog.WarnLevel, fields).Msg(msg)
}

func (l ZerologLogger) Error(ctx context.Context, err error, msg string) {
	l.newEvent(ctx, zerolog.ErrorLevel, nil).Err(err).Msg(msg)
}

func (l ZerologLogger) ErrorW(ctx context.Context, err error, msg string, fields map[string]any) {
	l.newEvent(ctx, zerolog.ErrorLevel, fields).Err(err).Msg(msg)
}

type (
	contextKey    struct{}
	contextFields struct{ fields map[string]any }
)

// ContextWithFields injects fields into ctx so they appear on every log line
// that passes this context. Fields are merged — later calls add to existing ones.
func ContextWithFields(ctx context.Context, fields map[string]any) context.Context {
	merged := make(map[string]any)
	if existing, ok := ctx.Value(contextKey{}).(contextFields); ok {
		for k, v := range existing.fields {
			merged[k] = v
		}
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, contextKey{}, contextFields{fields: merged})
}

func (l ZerologLogger) newEvent(ctx context.Context, level zerolog.Level, fields map[string]any) *zerolog.Event {
	event := applyContextFields(ctx, l.root.WithLevel(level))
	if len(fields) > 0 {
		event.Fields(fields)
	}
	return event.Caller(l.callerSkip)
}

func applyContextFields(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if ctx == nil {
		return event
	}
	if cf, ok := ctx.Value(contextKey{}).(contextFields); ok {
		event = event.Fields(cf.fields)
	}
	return event
}

func toZerologLevel(level Level) zerolog.Level {
	zl, err := zerolog.ParseLevel(strings.ToLower(string(level)))
	if err != nil {
		return zerolog.InfoLevel
	}
	return zl
}
