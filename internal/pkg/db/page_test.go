package db_test

import (
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/db"
)

func TestPage_Offset(t *testing.T) {
	tests := []struct {
		name string
		page db.Page
		want int
	}{
		{name: "first page", page: db.Page{Number: 1, Size: 20}, want: 0},
		{name: "second page", page: db.Page{Number: 2, Size: 20}, want: 20},
		{name: "third page size 10", page: db.Page{Number: 3, Size: 10}, want: 20},
		{name: "zero number", page: db.Page{Number: 0, Size: 20}, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.page.Offset()
			if got != tc.want {
				t.Errorf("Offset() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPage_Clamp(t *testing.T) {
	tests := []struct {
		name    string
		page    db.Page
		maxSize int
		want    db.Page
	}{
		{
			name:    "zero values get defaults",
			page:    db.Page{Number: 0, Size: 0},
			maxSize: 0,
			want:    db.Page{Number: 1, Size: 20},
		},
		{
			name:    "size exceeds max",
			page:    db.Page{Number: 1, Size: 100},
			maxSize: 50,
			want:    db.Page{Number: 1, Size: 50},
		},
		{
			name:    "already valid",
			page:    db.Page{Number: 2, Size: 15},
			maxSize: 100,
			want:    db.Page{Number: 2, Size: 15},
		},
		{
			name:    "max size zero — no cap",
			page:    db.Page{Number: 1, Size: 200},
			maxSize: 0,
			want:    db.Page{Number: 1, Size: 200},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.page.Clamp(tc.maxSize)
			if got != tc.want {
				t.Errorf("Clamp() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestNewPageResult(t *testing.T) {
	tests := []struct {
		name  string
		page  db.Page
		total int
		want  db.PageResult
	}{
		{
			name:  "exact fit",
			page:  db.Page{Number: 1, Size: 10},
			total: 10,
			want:  db.PageResult{Page: 1, Size: 10, Total: 10, TotalPages: 1},
		},
		{
			name:  "extra items on last page",
			page:  db.Page{Number: 1, Size: 10},
			total: 25,
			want:  db.PageResult{Page: 1, Size: 10, Total: 25, TotalPages: 3},
		},
		{
			name:  "zero total",
			page:  db.Page{Number: 1, Size: 20},
			total: 0,
			want:  db.PageResult{Page: 1, Size: 20, Total: 0, TotalPages: 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := db.NewPageResult(tc.page, tc.total)
			if got != tc.want {
				t.Errorf("NewPageResult() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
