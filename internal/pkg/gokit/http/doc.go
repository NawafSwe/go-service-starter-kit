// Package http provides go-kit HTTP transport helpers.
//
// MakeHTTPHandler assembles a go-kit HTTP server from an endpoint, codec,
// error encoder, and logger — with any number of endpoint middlewares applied:
//
//	handler := gohttp.MakeHTTPHandler(
//	    ep,
//	    codec,         // implements Decoder + Encoder
//	    codec,
//	    errorEncoder,
//	    lgr,
//	    middleware.TimeoutMiddleware(5*time.Second),
//	    middleware.RateLimit(time.Minute, 100),
//	)
//	router.Handle("/items", handler).Methods(http.MethodGet)
//
// LogErrorHandler bridges go-kit's transport.ErrorHandler to the service
// logger. Context-cancelled errors are swallowed since they are expected
// during graceful shutdown.
package http
