package example

import (
	"context"

	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
)

type (
	ListRequest  struct{}
	ListResponse struct {
		Examples []domain.Example
	}
)

type listRepository interface {
	List(ctx context.Context) ([]domain.Example, error)
}

// ListHandler handles the list-examples use case.
type ListHandler struct {
	repo listRepository
}

func NewListHandler(repo listRepository) ListHandler {
	return ListHandler{repo: repo}
}

func (h ListHandler) Handle(ctx context.Context, _ ListRequest) (ListResponse, error) {
	examples, err := h.repo.List(ctx)
	if err != nil {
		return ListResponse{}, err
	}
	return ListResponse{Examples: examples}, nil
}
