package v1

import (
	"encoding/json"
	"fmt"

	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
)

// CreateExampleCommandPayload is the transport-level wire format for the
// "example.create" message, matching the AsyncAPI CreateExampleCommandPayload schema.
type CreateExampleCommandPayload struct {
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		Name string `json:"name"`
	} `json:"data"`
}

// DecodeCreateExampleCommand decodes a JSON message payload into a business request.
func DecodeCreateExampleCommand(payload []byte) (any, error) {
	var p CreateExampleCommandPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("decode create-example command: %w", err)
	}
	if p.Data.Name == "" {
		return nil, fmt.Errorf("decode create-example command: name is required")
	}
	return example.CreateRequest{Name: p.Data.Name}, nil
}
