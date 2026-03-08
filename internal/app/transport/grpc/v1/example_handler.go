package v1

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	grpcv1 "github.com/nawafswe/go-service-starter-kit/api/proto/grpc/v1/gen"
)

// ExampleServiceHandler implements grpcv1.ExampleServiceServer by delegating each
// RPC method to a go-kit gRPC transport handler. This gives every method
// the same middleware chain (timeout, rate-limit, logging) as the HTTP transport.
type ExampleServiceHandler struct {
	grpcv1.UnimplementedExampleServiceServer
	createHandler grpctransport.Handler
	getHandler    grpctransport.Handler
	listHandler   grpctransport.Handler
	deleteHandler grpctransport.Handler
}

func NewExampleServiceHandler(
	createHandler grpctransport.Handler,
	getHandler grpctransport.Handler,
	listHandler grpctransport.Handler,
	deleteHandler grpctransport.Handler,
) *ExampleServiceHandler {
	return &ExampleServiceHandler{
		createHandler: createHandler,
		getHandler:    getHandler,
		listHandler:   listHandler,
		deleteHandler: deleteHandler,
	}
}

func (h *ExampleServiceHandler) CreateExample(ctx context.Context, req *grpcv1.CreateExampleRequest) (*grpcv1.ExampleResponse, error) {
	_, resp, err := h.createHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*grpcv1.ExampleResponse), nil
}

func (h *ExampleServiceHandler) GetExample(ctx context.Context, req *grpcv1.GetExampleRequest) (*grpcv1.ExampleResponse, error) {
	_, resp, err := h.getHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*grpcv1.ExampleResponse), nil
}

func (h *ExampleServiceHandler) ListExamples(ctx context.Context, req *grpcv1.ListExamplesRequest) (*grpcv1.ListExamplesResponse, error) {
	_, resp, err := h.listHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*grpcv1.ListExamplesResponse), nil
}

func (h *ExampleServiceHandler) DeleteExample(ctx context.Context, req *grpcv1.DeleteExampleRequest) (*grpcv1.DeleteExampleResponse, error) {
	_, resp, err := h.deleteHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*grpcv1.DeleteExampleResponse), nil
}
