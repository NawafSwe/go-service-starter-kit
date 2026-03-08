// Package consumer provides go-kit endpoint helpers for message consumers.
//
// MakeConsumerEndpoint wraps a go-kit endpoint with middlewares so that
// consumed messages flow through the same middleware chain (timeout,
// rate-limit, logging) as HTTP and gRPC requests:
//
//	ep := goconsumer.MakeConsumerEndpoint(
//	    endpointv1.MakeCreateExampleEndpoint(handler),
//	    middleware.TimeoutMiddleware(5*time.Second),
//	    middleware.RateLimit(time.Minute, 100),
//	)
//	resp, err := ep(ctx, request)
package consumer
