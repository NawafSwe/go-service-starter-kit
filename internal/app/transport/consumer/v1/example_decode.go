package v1

import (
	"encoding/json"
	"fmt"

	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
)

// createExampleCommandPayload matches the AsyncAPI CreateExampleCommandPayload schema.
type createExampleCommandPayload struct {
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		Name string `json:"name"`
	} `json:"data"`
}

// DecodeCreateExampleCommand decodes a JSON message payload into a CreateRequest.
func DecodeCreateExampleCommand(payload []byte) (example.CreateRequest, error) {
	var p createExampleCommandPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return example.CreateRequest{}, fmt.Errorf("decode create-example command: %w", err)
	}
	if p.Data.Name == "" {
		return example.CreateRequest{}, fmt.Errorf("decode create-example command: name is required")
	}
	return example.CreateRequest{Name: p.Data.Name}, nil
}
