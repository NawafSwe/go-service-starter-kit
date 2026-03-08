package db

// Page holds caller-supplied pagination input (1-indexed page number + page size).
type Page struct {
	Number int // 1-indexed; defaults to 1 when zero
	Size   int // items per page; defaults to 20 when zero
}

// Offset derives the SQL OFFSET value for this page.
func (p Page) Offset() int {
	if p.Number <= 1 {
		return 0
	}
	return (p.Number - 1) * p.Size
}

// Clamp enforces sane bounds:
//   - Number < 1  → 1
//   - Size < 1    → 20 (default)
//   - Size > maxSize (when maxSize > 0) → maxSize
func (p Page) Clamp(maxSize int) Page {
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = 20
	}
	if maxSize > 0 && p.Size > maxSize {
		p.Size = maxSize
	}
	return p
}

// PageResult is the computed pagination metadata returned in list responses.
type PageResult struct {
	Page       int
	Size       int
	Total      int
	TotalPages int
}

// NewPageResult computes PageResult from a clamped Page and a total count.
func NewPageResult(p Page, total int) PageResult {
	totalPages := 0
	if p.Size > 0 {
		totalPages = (total + p.Size - 1) / p.Size
	}
	return PageResult{
		Page:       p.Number,
		Size:       p.Size,
		Total:      total,
		TotalPages: totalPages,
	}
}
