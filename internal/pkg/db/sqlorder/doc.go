// Package sqlorder provides a safe SQL ORDER BY builder that sanitises
// column names to prevent SQL injection.
//
//	order := sqlorder.OrderBy{Column: req.Sort, Direction: sqlorder.Desc}
//	query := "SELECT ... ORDER BY " + order.SQL()
package sqlorder
