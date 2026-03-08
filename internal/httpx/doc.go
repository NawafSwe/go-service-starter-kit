// Package httpx provides a resilient HTTP client with circuit breaking,
// exponential back-off retries, and built-in OpenTelemetry tracing and metrics.
//
// # Quick start
//
//	client, err := httpx.New(
//	    httpx.Config{
//	        Name:         "payment-service",
//	        BaseURL:      "https://api.example.com",
//	        Timeout:      5 * time.Second,
//	        MaxRetries:   3,
//	        RetryWaitMin: 100 * time.Millisecond,
//	        RetryWaitMax: 1 * time.Second,
//	        CircuitBreaker: httpx.CircuitBreakerConfig{
//	            MaxRequests: 5,
//	            Timeout:     30 * time.Second,
//	            Threshold:   5,
//	        },
//	    },
//	    &http.Client{},
//	    httpx.WithMeter(meter),             // optional
//	    httpx.WithTracerProvider(tp),       // optional
//	)
//
//	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/items", nil)
//	resp, err := client.Do(ctx, req)
//
// # Retry behaviour
//
// Retries are triggered on transport errors or 5xx responses. The backoff
// uses exponential jitter: base doubles each attempt, capped at RetryWaitMax,
// with random jitter up to the base duration. Set RetryWaitMin ≤ 0 to disable
// the wait between retries.
//
// # Circuit breaker
//
// The circuit opens after ConsecutiveFailures ≥ Threshold. Set Threshold to 0
// to disable the circuit breaker entirely.
//
// # Observability
//
// Every call increments the http_client_requests_total counter with attributes:
// dependency, circuit_breaker, method, url, and result (success|error).
// A client-side span is started for each call.
package httpx
