// Package grpc provides go-kit gRPC transport helpers.
//
// MakeGRPCHandler assembles a go-kit gRPC server handler from an endpoint,
// decode/encode functions, and logger — with any number of endpoint middlewares:
//
//	handler := gogrpc.MakeGRPCHandler(
//	    ep,
//	    decodeCreateRequest,
//	    encodeCreateResponse,
//	    lgr,
//	    middleware.TimeoutMiddleware(5*time.Second),
//	    middleware.RateLimit(time.Minute, 100),
//	)
package grpc
