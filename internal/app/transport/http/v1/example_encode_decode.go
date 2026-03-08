package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
	transporterrors "github.com/nawafswe/go-service-starter-kit/internal/httperrors"
)

// exampleJSON is the wire representation of a domain.Example.
type exampleJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toExampleJSON(e domain.Example) exampleJSON {
	return exampleJSON{
		ID:        e.ID.String(),
		Name:      e.Name,
		CreatedAt: e.CreatedAt.String(),
		UpdatedAt: e.UpdatedAt.String(),
	}
}

// ---- Create ----

type CreateExampleEncoderDecoder struct{}

func (CreateExampleEncoderDecoder) Decode(_ context.Context, r *http.Request) (any, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, transporterrors.NewBadRequestError("invalid request body")
	}
	if body.Name == "" {
		return nil, transporterrors.NewBadRequestError("name is required")
	}
	return example.CreateRequest{Name: body.Name}, nil
}

func (CreateExampleEncoderDecoder) Encode(_ context.Context, w http.ResponseWriter, response any) error {
	r, ok := response.(example.CreateResponse)
	if !ok {
		return fmt.Errorf("expected %T, got %T", example.CreateResponse{}, response)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(toExampleJSON(r.Example))
}

// ---- Get ----

type GetExampleEncoderDecoder struct{}

func (GetExampleEncoderDecoder) Decode(_ context.Context, r *http.Request) (any, error) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		return nil, transporterrors.NewBadRequestError("invalid id")
	}
	return example.GetRequest{ID: id}, nil
}

func (GetExampleEncoderDecoder) Encode(_ context.Context, w http.ResponseWriter, response any) error {
	r, ok := response.(example.GetResponse)
	if !ok {
		return fmt.Errorf("expected %T, got %T", example.GetResponse{}, response)
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(toExampleJSON(r.Example))
}

// ---- List ----

type ListExamplesEncoderDecoder struct{}

func (ListExamplesEncoderDecoder) Decode(_ context.Context, _ *http.Request) (any, error) {
	return example.ListRequest{}, nil
}

func (ListExamplesEncoderDecoder) Encode(_ context.Context, w http.ResponseWriter, response any) error {
	r, ok := response.(example.ListResponse)
	if !ok {
		return fmt.Errorf("expected %T, got %T", example.ListResponse{}, response)
	}
	items := make([]exampleJSON, len(r.Examples))
	for i, ex := range r.Examples {
		items[i] = toExampleJSON(ex)
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(items)
}

// ---- Delete ----

type DeleteExampleEncoderDecoder struct{}

func (DeleteExampleEncoderDecoder) Decode(_ context.Context, r *http.Request) (any, error) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		return nil, transporterrors.NewBadRequestError("invalid id")
	}
	return example.DeleteRequest{ID: id}, nil
}

func (DeleteExampleEncoderDecoder) Encode(_ context.Context, w http.ResponseWriter, _ any) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
