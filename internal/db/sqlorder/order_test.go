package sqlorder_test

import (
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/db/sqlorder"
)

func TestOrderBy_SQL(t *testing.T) {
	tests := []struct {
		name  string
		order sqlorder.OrderBy
		want  string
	}{
		{
			name:  "valid asc",
			order: sqlorder.OrderBy{Column: "name", Direction: sqlorder.Asc},
			want:  "name ASC",
		},
		{
			name:  "valid desc",
			order: sqlorder.OrderBy{Column: "created_at", Direction: sqlorder.Desc},
			want:  "created_at DESC",
		},
		{
			name:  "invalid direction defaults to DESC",
			order: sqlorder.OrderBy{Column: "id", Direction: "RANDOM"},
			want:  "id DESC",
		},
		{
			name:  "empty column falls back to default",
			order: sqlorder.OrderBy{Column: "", Direction: sqlorder.Desc},
			want:  "created_at DESC",
		},
		{
			name:  "SQL injection chars stripped",
			order: sqlorder.OrderBy{Column: "id; DROP TABLE", Direction: sqlorder.Desc},
			want:  "id DESC",
		},
		{
			name:  "uppercase letters stripped",
			order: sqlorder.OrderBy{Column: "Name", Direction: sqlorder.Asc},
			want:  "ame ASC",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.order.SQL()
			if got != tc.want {
				t.Errorf("SQL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDefaultOrder(t *testing.T) {
	got := sqlorder.DefaultOrder.SQL()
	if got != "created_at DESC" {
		t.Errorf("DefaultOrder.SQL() = %q, want %q", got, "created_at DESC")
	}
}
