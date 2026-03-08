package example

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
	examplerepo "github.com/nawafswe/go-service-starter-kit/internal/app/repositories/example"
)

//go:generate go tool mockgen -source=${GOFILE} -destination=mock/${GOFILE} -package=mock

type (
	CreateRequest struct {
		Name string
	}
	CreateResponse struct {
		Example domain.Example
	}
)

type exampleRepository interface {
	Create(ctx context.Context, params examplerepo.CreateParams) (domain.Example, error)
}

// CreateHandler handles the create-example use case.
type CreateHandler struct {
	repo exampleRepository
}

func NewCreateHandler(repo exampleRepository) CreateHandler {
	return CreateHandler{repo: repo}
}

func (h CreateHandler) Handle(ctx context.Context, req CreateRequest) (CreateResponse, error) {
	if req.Name == "" {
		return CreateResponse{}, fmt.Errorf("%w: name is required", domain.ErrInvalidRequest)
	}
	example, err := h.repo.Create(ctx, examplerepo.CreateParams{
		ID:   uuid.New(),
		Name: req.Name,
	})
	if err != nil {
		return CreateResponse{}, err
	}
	return CreateResponse{Example: example}, nil
}
