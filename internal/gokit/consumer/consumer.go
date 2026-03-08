package consumer

import "github.com/go-kit/kit/endpoint"

// MakeConsumerEndpoint wraps a go-kit endpoint with middlewares for use in message consumers.
// The returned endpoint applies the middleware chain in order, matching the behaviour
// of the HTTP and gRPC transport factories.
func MakeConsumerEndpoint(
	ep endpoint.Endpoint,
	middlewares ...endpoint.Middleware,
) endpoint.Endpoint {
	for _, mw := range middlewares {
		ep = mw(ep)
	}
	return ep
}
