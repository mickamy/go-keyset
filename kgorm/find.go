package kgorm

import (
	"gorm.io/gorm"

	"github.com/mickamy/go-keyset"
)

// FindPage executes the provided GORM query into out and normalizes the
// result order for keyset pagination. If page.Dir is DirPrev, it reverses
// the slice in-place via keyset.NormalizePageResult so that the final order
// matches the requested display order.
//
// Usage:
//
//	var items []Item
//	tx := FindPage(
//	    PageByTimeAndID(db, page, keyset.Descending, "created_at", "id"),
//	    page, &items,
//	)
//	if tx.Error != nil { ... }
func FindPage[T any](q *gorm.DB, page keyset.Page, out *[]T) *gorm.DB {
	tx := q.Find(out)
	if tx.Error == nil && out != nil {
		*out = keyset.NormalizePageResult(page, *out)
	}
	return tx
}
