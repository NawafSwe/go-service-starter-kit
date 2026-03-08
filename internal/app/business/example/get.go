package example

import (
	"context"

	"github.com/google/uuid"
	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
)

type (
	GetRequest struct {
		ID uuid.UUID
	}
	GetResponse struct {
		Example domain.Example
	}
)

type getRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Example, error)
}

// GetHandler handles the get-example use case.
type GetHandler struct {
	repo getRepository
}

func NewGetHandler(repo getRepository) GetHandler {
	return GetHandler{repo: repo}
}

func (h GetHandler) Handle(ctx context.Context, req GetRequest) (GetResponse, error) {
	example, err := h.repo.GetByID(ctx, req.ID)
	if err != nil {
		return GetResponse{}, err
	}
	return GetResponse{Example: example}, nil
}
