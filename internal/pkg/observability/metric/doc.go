// Package metric provides a thin Reporter wrapper over OpenTelemetry metrics.
//
// Create a Reporter once at startup and inject it into application code:
//
//	reporter, err := metric.NewReporter(otelMeter)
//
// Record each processed request with low-cardinality attributes:
//
//	reporter.RecordRequest(ctx, map[string]any{
//	    "endpoint": "/items",
//	    "method":   "GET",
//	    "status":   200,
//	    "success":  true,
//	})
//
// Supported attribute value types: string, bool, int, int64.
// Unknown types are silently skipped to prevent accidental panics.
//
// Avoid high-cardinality attributes (user IDs, session IDs, full URLs)
// to prevent metric explosion in your observability backend.
package metric
