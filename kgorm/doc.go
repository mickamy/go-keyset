// Package kgorm provides GORM integration for the go-keyset library.
//
// It offers composable GORM scopes for building stable, bidirectional
// keyset pagination queries.
//
// Example:
//
//	import (
//	    "github.com/mickamy/go-keyset"
//	    "github.com/mickamy/go-keyset/kgorm"
//	)
//
//	var page = keyset.Page{
//	    Limit: 10,
//	    Dir:   keyset.DirNext,
//	}
//
//	var posts []Post
//
//	tx := kgorm.FindPage(
//	    kgorm.PageByTimeAndID(
//	        db.Model(&Post{}),
//	        page,
//	        keyset.Descending,
//	        "created_at",
//	        "id",
//	    ),
//	    page,
//	    &posts,
//	)
//
// The generated SQL will be stable across inserts or deletions, for example:
//
//	SELECT * FROM posts
//	WHERE (created_at < $1) OR (created_at = $1 AND id < $2)
//	ORDER BY created_at DESC, id DESC
//	LIMIT $3;
//
// This package is a thin layer around GORM that delegates all pagination
// logic to the core keyset package, providing only ORM-specific composition
// and execution helpers.
package kgorm
