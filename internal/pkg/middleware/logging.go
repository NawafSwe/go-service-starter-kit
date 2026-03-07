package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ggwhite/go-masker"
	"github.com/go-kit/kit/endpoint"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

// LoggingPubSubMiddleware logs pub/sub message processing with masking.
// Attaches topic and transport attributes to every log line.
func LoggingPubSubMiddleware(lgr logger.ZerologLogger, topic string) endpoint.Middleware {
	lgr = lgr.WithFields(map[string]any{
		"topic":     topic,
		"transport": "pubsub",
	})
	return loggingMiddleware(lgr)
}

// LoggingHTTPMiddleware logs HTTP endpoint processing with masking.
func LoggingHTTPMiddleware(lgr logger.ZerologLogger, url, method string) endpoint.Middleware {
	lgr = lgr.WithFields(map[string]any{
		"transport": "http",
		"url":       url,
		"method":    method,
	})
	return loggingMiddleware(lgr)
}

// LoggingGRPCMiddleware logs gRPC endpoint processing with masking.
func LoggingGRPCMiddleware(lgr logger.ZerologLogger, function string) endpoint.Middleware {
	lgr = lgr.WithFields(map[string]any{
		"transport": "grpc",
		"function":  function,
	})
	return loggingMiddleware(lgr)
}

// loggingMiddleware is the shared implementation — logs duration, masked request, masked response.
func loggingMiddleware(lgr logger.ZerologLogger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request any) (any, error) {
			begin := time.Now()
			response, err := next(ctx, request)

			attrs, attErr := buildLogAttrs(request, response, time.Since(begin))
			if attErr != nil {
				lgr.Error(ctx, attErr, "failed to build log attributes")
			}
			if err == nil {
				lgr.InfoFields(ctx, "processed successfully", attrs)
			} else {
				lgr.ErrorFields(ctx, err, "processing failed", attrs)
			}
			return response, err
		}
	}
}

// buildLogAttrs marshals request and response with sensitive-field masking applied.
func buildLogAttrs(request, response any, duration time.Duration) (map[string]any, error) {
	attrs := map[string]any{"duration_ms": duration.Milliseconds()}

	reqBytes, err := maskedJSON(request)
	if err != nil {
		return attrs, fmt.Errorf("mask request: %w", err)
	}
	attrs["request"] = json.RawMessage(reqBytes)

	respBytes, err := maskedJSON(response)
	if err != nil {
		return attrs, fmt.Errorf("mask response: %w", err)
	}
	attrs["response"] = json.RawMessage(respBytes)

	return attrs, nil
}

// maskedJSON marshals v after applying go-masker field masking.
func maskedJSON(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	masked, err := masker.Struct(v)
	if err != nil {
		return nil, fmt.Errorf("masker.Struct: %w", err)
	}
	return json.Marshal(masked)
}
