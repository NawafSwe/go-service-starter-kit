package domain

import (
	"time"

	"github.com/google/uuid"
)

// Example is a placeholder domain entity.
// Replace this with your own domain models.
type Example struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
