// Package grpcx provides a resilient gRPC invoker with circuit breaking,
// exponential back-off retries, and built-in OpenTelemetry tracing and metrics.
//
// # Quick start
//
//	conn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
//
//	client, err := grpcx.New(
//	    grpcx.Config{
//	        Name:         "inventory-service",
//	        Address:      "localhost:50051",
//	        Timeout:      5 * time.Second,
//	        MaxRetries:   3,
//	        RetryWaitMin: 50 * time.Millisecond,
//	        RetryWaitMax: 500 * time.Millisecond,
//	        CircuitBreaker: grpcx.CircuitBreakerConfig{
//	            MaxRequests: 5,
//	            Timeout:     30 * time.Second,
//	            Threshold:   5,
//	        },
//	    },
//	    conn,
//	    meter,
//	    tracerProvider,
//	)
//
//	err = client.Invoke(ctx, "/inventory.v1.InventoryService/GetItem", req, &resp)
//
// # Retry behaviour
//
// Retries on any non-nil error except gobreaker.ErrOpenState or ErrTooManyRequests
// (those break the loop immediately). Jitter backoff applies between attempts.
//
// # Circuit breaker
//
// The circuit opens after ConsecutiveFailures ≥ Threshold. Set Threshold to 0
// to disable the circuit breaker entirely.
//
// # Observability
//
// Every call increments the grpc_client_requests_total counter with attributes:
// dependency, circuit_breaker, method, and result (success|error).
// A SpanKindClient span is started for each call.
package grpcx
