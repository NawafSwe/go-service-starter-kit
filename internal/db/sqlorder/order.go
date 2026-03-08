package sqlorder

// SortDirection is either ascending or descending.
type SortDirection string

const (
	Asc  SortDirection = "ASC"
	Desc SortDirection = "DESC"
)

// OrderBy describes a single SQL ORDER BY clause.
type OrderBy struct {
	Column    string
	Direction SortDirection
}

// SQL returns a safe, validated ORDER BY fragment (e.g. "created_at DESC").
// The column is sanitized — only lowercase letters, digits, and underscores are
// allowed; any invalid value falls back to "created_at DESC".
func (o OrderBy) SQL() string {
	col := sanitizeColumn(o.Column)
	if col == "" {
		return "created_at DESC"
	}
	dir := o.Direction
	if dir != Asc && dir != Desc {
		dir = Desc
	}
	return col + " " + string(dir)
}

// DefaultOrder is the standard ordering for list endpoints — newest first.
var DefaultOrder = OrderBy{Column: "created_at", Direction: Desc}

// sanitizeColumn strips any character that isn't a lowercase letter, digit, or underscore.
// This prevents SQL injection via column names that are not positional parameters.
func sanitizeColumn(col string) string {
	out := make([]byte, 0, len(col))
	for i := range len(col) {
		c := col[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			out = append(out, c)
		}
	}
	return string(out)
}
