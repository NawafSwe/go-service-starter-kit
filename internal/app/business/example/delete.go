package example

import (
	"context"

	"github.com/google/uuid"
)

type (
	DeleteRequest struct {
		ID uuid.UUID
	}
	DeleteResponse struct{}
)

type deleteRepository interface {
	Delete(ctx context.Context, id uuid.UUID) error
}

// DeleteHandler handles the delete-example use case.
type DeleteHandler struct {
	repo deleteRepository
}

func NewDeleteHandler(repo deleteRepository) DeleteHandler {
	return DeleteHandler{repo: repo}
}

func (h DeleteHandler) Handle(ctx context.Context, req DeleteRequest) (DeleteResponse, error) {
	if err := h.repo.Delete(ctx, req.ID); err != nil {
		return DeleteResponse{}, err
	}
	return DeleteResponse{}, nil
}
