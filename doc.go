// Package keyset provides a minimal and type-safe foundation
// for keyset pagination in Go.
//
// Unlike offset-based pagination, keyset pagination uses stable
// sort keys (e.g., IDs or timestamps) to achieve consistent and
// efficient page navigation.
//
// Example usage (with GORM adapter):
//
//	import (
//	    "github.com/mickamy/go-keyset"
//	    "github.com/mickamy/go-keyset/kgorm"
//	)
//
//	var page = keyset.Page{Limit: 10, Dir: keyset.DirNext}
//	var posts []Post
//
//	db = kgorm.PageByTimeAndID(db, page, keyset.Descending, "created_at", "id")
//	db.Find(&posts)
//	keyset.NormalizePageResult(page, &posts)
//
// Core features:
//   - Stable keyset pagination with bidirectional navigation
//   - Opaque cursor encoding (int64, time, or composite time+id)
//   - Direction- and order-aware SQL helpers
//
// For a practical example, see examples/kgorm.
package keyset
