package metric

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	otelmeter "go.opentelemetry.io/otel/metric"
)

const (
	requestCounterName = "app_requests_total"
	requestCounterUnit = "1"
)

// Reporter holds pre-initialised OTel instruments for common service metrics.
// Extend it with your own counters and histograms as the service grows.
type Reporter struct {
	requestCounter otelmeter.Int64Counter
}

// NewReporter registers the service-level metric instruments on the provided meter.
func NewReporter(m otelmeter.Meter) (*Reporter, error) {
	counter, err := m.Int64Counter(
		requestCounterName,
		otelmeter.WithDescription("Total number of processed requests"),
		otelmeter.WithUnit(requestCounterUnit),
	)
	if err != nil {
		return nil, fmt.Errorf("metric: register %s: %w", requestCounterName, err)
	}
	return &Reporter{requestCounter: counter}, nil
}

// RecordRequest increments the request counter with the provided attributes.
// Avoid high-cardinality attributes (user IDs, session IDs) to prevent metric explosion.
func (r *Reporter) RecordRequest(ctx context.Context, attrs map[string]any) {
	r.requestCounter.Add(ctx, 1, otelmeter.WithAttributes(toKeyValues(attrs)...))
}

// toKeyValues converts a map to OTel attribute.KeyValue pairs.
// Supported types: string, bool, int. Unknown types are silently skipped.
func toKeyValues(attrs map[string]any) []attribute.KeyValue {
	kvs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		switch val := v.(type) {
		case string:
			kvs = append(kvs, attribute.String(k, val))
		case bool:
			kvs = append(kvs, attribute.Bool(k, val))
		case int:
			kvs = append(kvs, attribute.Int64(k, int64(val)))
		case int64:
			kvs = append(kvs, attribute.Int64(k, val))
		}
	}
	return kvs
}
