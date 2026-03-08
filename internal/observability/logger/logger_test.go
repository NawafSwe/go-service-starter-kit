package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer) logger.ZerologLogger {
	return logger.NewLogger(logger.DebugLevel, "test-app", "1.0.0", "test").WithOutput(buf)
}

func TestLoggerSetup(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(buf *bytes.Buffer) logger.ZerologLogger
		contains []string
	}{
		{
			name: "basic logger writes output",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return logger.NewLogger(logger.InfoLevel, "svc", "1.0", "prod").WithOutput(buf)
			},
			contains: []string{"ping"},
		},
		{
			name: "WithFields adds fields",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return newTestLogger(buf).WithFields(map[string]any{"component": "worker"})
			},
			contains: []string{"component"},
		},
		{
			name: "WithHost adds host",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return newTestLogger(buf).WithHost("my-host")
			},
			contains: []string{"my-host"},
		},
		{
			name: "WithOutput directs output",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return logger.NewLogger(logger.DebugLevel, "app", "1.0", "test").WithOutput(buf)
			},
			contains: []string{"test output"},
		},
		{
			name: "ContextWithFields injects fields",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return newTestLogger(buf)
			},
			contains: []string{"req-123"},
		},
		{
			name: "ContextWithFields merges fields",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return newTestLogger(buf)
			},
			contains: []string{`"a"`, `"b"`},
		},
		{
			name: "output is valid JSON",
			setup: func(buf *bytes.Buffer) logger.ZerologLogger {
				return newTestLogger(buf)
			},
			contains: []string{"json check"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lgr := tc.setup(&buf)

			switch tc.name {
			case "ContextWithFields injects fields":
				ctx := logger.ContextWithFields(context.Background(), map[string]any{"request_id": "req-123"})
				lgr.Info(ctx, "handled")
			case "ContextWithFields merges fields":
				ctx := logger.ContextWithFields(context.Background(), map[string]any{"a": "1"})
				ctx = logger.ContextWithFields(ctx, map[string]any{"b": "2"})
				lgr.Info(ctx, "merged")
			case "output is valid JSON":
				lgr.Info(context.Background(), "json check")
				var m map[string]any
				require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
			case "WithOutput directs output":
				lgr.Info(context.Background(), "test output")
			default:
				lgr.Info(context.Background(), "ping")
			}

			out := buf.String()
			assert.NotEmpty(t, out)
			for _, s := range tc.contains {
				assert.True(t, strings.Contains(out, s), "output missing %q: %s", s, out)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		log      func(lgr logger.ZerologLogger)
		contains string
	}{
		{name: "debug", log: func(lgr logger.ZerologLogger) { lgr.Debug(context.Background(), "debug message") }, contains: "debug message"},
		{name: "info", log: func(lgr logger.ZerologLogger) { lgr.Info(context.Background(), "info message") }, contains: "info message"},
		{name: "warn", log: func(lgr logger.ZerologLogger) { lgr.Warn(context.Background(), "warn message") }, contains: "warn message"},
		{name: "error", log: func(lgr logger.ZerologLogger) { lgr.Error(context.Background(), errors.New("boom"), "error message") }, contains: "boom"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lgr := newTestLogger(&buf)
			tc.log(lgr)
			assert.Contains(t, buf.String(), tc.contains)
		})
	}
}

func TestWithLevel_FiltersBelow(t *testing.T) {
	var buf bytes.Buffer
	lgr := newTestLogger(&buf).WithLevel(logger.WarnLevel)
	lgr.Debug(context.Background(), "this should be filtered")
	lgr.Info(context.Background(), "this too")

	assert.Empty(t, buf.String())
}

func TestLogWithFields(t *testing.T) {
	tests := []struct {
		name     string
		log      func(lgr logger.ZerologLogger)
		contains string
	}{
		{
			name: "InfoW",
			log: func(lgr logger.ZerologLogger) {
				lgr.InfoW(context.Background(), "with fields", map[string]any{"key": "value"})
			},
			contains: "value",
		},
		{
			name: "DebugW",
			log: func(lgr logger.ZerologLogger) {
				lgr.DebugW(context.Background(), "debug fields", map[string]any{"dbg": "yes"})
			},
			contains: "dbg",
		},
		{
			name: "WarnW",
			log: func(lgr logger.ZerologLogger) {
				lgr.WarnW(context.Background(), "warn fields", map[string]any{"severity": "medium"})
			},
			contains: "medium",
		},
		{
			name: "ErrorW",
			log: func(lgr logger.ZerologLogger) {
				lgr.ErrorW(context.Background(), errors.New("err"), "fields message", map[string]any{"trace": "abc"})
			},
			contains: "abc",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lgr := newTestLogger(&buf)
			tc.log(lgr)
			assert.Contains(t, buf.String(), tc.contains)
		})
	}
}

func TestWithPrettyOutput(t *testing.T) {
	lgr := logger.NewLogger(logger.DebugLevel, "test-app", "1.0.0", "test").WithPrettyOutput()
	// WithPrettyOutput writes to os.Stderr; verify it doesn't panic.
	assert.NotPanics(t, func() {
		lgr.Info(context.Background(), "pretty output test")
	})
}

func TestWithCallerSkip(t *testing.T) {
	var buf bytes.Buffer
	lgr := newTestLogger(&buf).WithCallerSkip(3)
	lgr.Info(context.Background(), "caller skip test")
	assert.Contains(t, buf.String(), "caller skip test")
}

func TestNilContext(t *testing.T) {
	var buf bytes.Buffer
	lgr := newTestLogger(&buf)
	//nolint:staticcheck // intentionally passing nil context to test the nil guard
	lgr.Info(nil, "nil context test")
	assert.Contains(t, buf.String(), "nil context test")
}

func TestToZerologLevel_InvalidLevel(t *testing.T) {
	// Invalid level defaults to info — debug messages should be filtered.
	var buf bytes.Buffer
	lgr := logger.NewLogger("invalid-level", "app", "1.0", "test").WithOutput(&buf)
	lgr.Debug(context.Background(), "should be filtered")
	assert.Empty(t, buf.String())

	lgr.Info(context.Background(), "should appear")
	assert.Contains(t, buf.String(), "should appear")
}
