// Package middleware provides transport-agnostic middleware for HTTP, gRPC,
// and pub/sub (Kafka, PubSub, etc.) transports.
//
// # JWT Authentication
//
// Three variants are provided for each transport — required, optional, and mock:
//
//   - AuthRequired / GRPCAuthRequired / ConsumerAuthRequired:
//     Reject the request when no valid JWT is present.
//
//   - AuthOptional / GRPCAuthOptional / ConsumerAuthOptional:
//     Pass unauthenticated requests through; validate when a token is present.
//
//   - AuthMock / GRPCAuthMock / ConsumerAuthMock:
//     Inject a fixed user from config — for local development only.
//
// # Rate limiting and timeouts
//
// TimeoutMiddleware wraps a go-kit endpoint with a context deadline.
// RateLimit provides a simple sliding-window rate limiter.
// Both are transport-agnostic and compose with any endpoint.Middleware chain.
//
// # Logging
//
// LoggingHTTPMiddleware, LoggingGRPCMiddleware, and LoggingPubSubMiddleware
// attach transport-specific fields to every log line and log the duration,
// masked request, and masked response for each processed message.
package middleware
