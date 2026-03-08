package v1

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
)

type createExampleHandler interface {
	Handle(ctx context.Context, req example.CreateRequest) (example.CreateResponse, error)
}

func MakeCreateExampleEndpoint(h createExampleHandler) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(example.CreateRequest)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", example.CreateRequest{}, request)
		}
		return h.Handle(ctx, req)
	}
}

type getExampleHandler interface {
	Handle(ctx context.Context, req example.GetRequest) (example.GetResponse, error)
}

func MakeGetExampleEndpoint(h getExampleHandler) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(example.GetRequest)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", example.GetRequest{}, request)
		}
		return h.Handle(ctx, req)
	}
}

type listExamplesHandler interface {
	Handle(ctx context.Context, req example.ListRequest) (example.ListResponse, error)
}

func MakeListExamplesEndpoint(h listExamplesHandler) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(example.ListRequest)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", example.ListRequest{}, request)
		}
		return h.Handle(ctx, req)
	}
}

type deleteExampleHandler interface {
	Handle(ctx context.Context, req example.DeleteRequest) (example.DeleteResponse, error)
}

func MakeDeleteExampleEndpoint(h deleteExampleHandler) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(example.DeleteRequest)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", example.DeleteRequest{}, request)
		}
		return h.Handle(ctx, req)
	}
}
