package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	grpcv1 "github.com/nawafswe/go-service-starter-kit/api/proto/grpc/v1/gen"
	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ---- Encode helpers ----

func toProto(e domain.Example) *grpcv1.Example {
	return &grpcv1.Example{
		Id:        e.ID.String(),
		Name:      e.Name,
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}
}

// ---- Create ----

func DecodeCreateExampleRequest(_ context.Context, grpcReq any) (any, error) {
	req, ok := grpcReq.(*grpcv1.CreateExampleRequest)
	if !ok {
		return nil, fmt.Errorf("expected *grpcv1.CreateExampleRequest, got %T", grpcReq)
	}
	return example.CreateRequest{Name: req.GetName()}, nil
}

func EncodeCreateExampleResponse(_ context.Context, response any) (any, error) {
	resp, ok := response.(example.CreateResponse)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", example.CreateResponse{}, response)
	}
	return &grpcv1.ExampleResponse{Example: toProto(resp.Example)}, nil
}

// ---- Get ----

func DecodeGetExampleRequest(_ context.Context, grpcReq any) (any, error) {
	req, ok := grpcReq.(*grpcv1.GetExampleRequest)
	if !ok {
		return nil, fmt.Errorf("expected *grpcv1.GetExampleRequest, got %T", grpcReq)
	}
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}
	return example.GetRequest{ID: id}, nil
}

func EncodeGetExampleResponse(_ context.Context, response any) (any, error) {
	resp, ok := response.(example.GetResponse)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", example.GetResponse{}, response)
	}
	return &grpcv1.ExampleResponse{Example: toProto(resp.Example)}, nil
}

// ---- List ----

func DecodeListExamplesRequest(_ context.Context, _ any) (any, error) {
	return example.ListRequest{}, nil
}

func EncodeListExamplesResponse(_ context.Context, response any) (any, error) {
	resp, ok := response.(example.ListResponse)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", example.ListResponse{}, response)
	}
	items := make([]*grpcv1.Example, len(resp.Examples))
	for i, e := range resp.Examples {
		items[i] = toProto(e)
	}
	return &grpcv1.ListExamplesResponse{Examples: items}, nil
}

// ---- Delete ----

func DecodeDeleteExampleRequest(_ context.Context, grpcReq any) (any, error) {
	req, ok := grpcReq.(*grpcv1.DeleteExampleRequest)
	if !ok {
		return nil, fmt.Errorf("expected *grpcv1.DeleteExampleRequest, got %T", grpcReq)
	}
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}
	return example.DeleteRequest{ID: id}, nil
}

func EncodeDeleteExampleResponse(_ context.Context, _ any) (any, error) {
	return &grpcv1.DeleteExampleResponse{}, nil
}
