package example

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
)

type exampleRecord struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func toDomain(r exampleRecord) domain.Example {
	return domain.Example{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// CreateParams holds the data required to insert a new example record.
type CreateParams struct {
	ID   uuid.UUID
	Name string
}

// Repository is the example domain repository.
type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r Repository) Create(ctx context.Context, params CreateParams) (domain.Example, error) {
	const q = `
		INSERT INTO examples (id, name, created_at, updated_at)
		VALUES (:id, :name, :created_at, :updated_at)
		RETURNING id, name, created_at, updated_at`

	values := map[string]any{
		"id":         params.ID,
		"name":       params.Name,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	rows, err := r.db.NamedQueryContext(ctx, q, values)
	if err != nil {
		return domain.Example{}, fmt.Errorf("example repository: create: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return domain.Example{}, fmt.Errorf("example repository: create: no row returned")
	}
	var rec exampleRecord
	if err := rows.StructScan(&rec); err != nil {
		return domain.Example{}, fmt.Errorf("example repository: create: scan: %w", err)
	}
	return toDomain(rec), nil
}

func (r Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Example, error) {
	const q = `SELECT id, name, created_at, updated_at FROM examples WHERE id = $1`

	var rec exampleRecord
	if err := r.db.GetContext(ctx, &rec, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Example{}, domain.ErrNotFound
		}
		return domain.Example{}, fmt.Errorf("example repository: get by id: %w", err)
	}
	return toDomain(rec), nil
}

func (r Repository) List(ctx context.Context) ([]domain.Example, error) {
	const q = `SELECT id, name, created_at, updated_at FROM examples ORDER BY created_at DESC`

	var records []exampleRecord
	if err := r.db.SelectContext(ctx, &records, q); err != nil {
		return nil, fmt.Errorf("example repository: list: %w", err)
	}
	examples := make([]domain.Example, len(records))
	for i, rec := range records {
		examples[i] = toDomain(rec)
	}
	return examples, nil
}

func (r Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM examples WHERE id = $1`

	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("example repository: delete: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("example repository: delete: rows affected: %w", err)
	}
	if count == 0 {
		return domain.ErrNotFound
	}
	return nil
}
