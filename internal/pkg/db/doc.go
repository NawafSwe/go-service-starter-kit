// Package db provides shared, database-agnostic pagination utilities
// used across all repository layers.
//
//	page := db.Page{Number: req.Page, Size: req.Size}.Clamp(100)
//	offset := page.Offset()
//	// SELECT ... LIMIT $1 OFFSET $2 — bind page.Size, offset
//
// Return PageResult in list responses so callers know total pages:
//
//	result := db.NewPageResult(page, totalCount)
//
// For SQL-specific ordering helpers see the [db/sqlorder] sub-package.
package db
