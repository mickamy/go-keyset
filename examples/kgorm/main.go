package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mickamy/go-keyset"
	"github.com/mickamy/go-keyset/examples/internal/model"
	"github.com/mickamy/go-keyset/kgorm"
)

const dsn = "postgres://postgres:password@localhost:5432/keyset?sslmode=disable"

func main() {
	ctx := context.Background()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	fmt.Println("=== Keyset pagination: created_at DESC, id DESC ===")

	// Start from the latest page (no cursor), limit 5
	page := keyset.Page{
		Limit: 5,
		Dir:   keyset.DirNext, // move forward
	}

	// Walk two "next" pages
	var cursor string
	for i := 0; i < 2; i++ {
		page.Cursor = cursor
		posts, next := fetchPageByTimeAndID(ctx, db, page, keyset.Descending)
		printPage(i+1, page.Dir, posts)
		cursor = next // carry forward
	}

	// Now go "prev" once from the current cursor
	page.Dir = keyset.DirPrev
	page.Cursor = cursor
	posts, prev := fetchPageByTimeAndID(ctx, db, page, keyset.Descending)
	printPage(1, page.Dir, posts)
	_ = prev // not used further; just showing the symmetric flow
}

// fetchPageByTimeAndID runs a single page query by (created_at, id)
// using DESC order, applies keyset windows, and returns results plus the
// cursor for the next request in the SAME direction.
// For DirNext: next cursor = last row boundary
// For DirPrev: next cursor = last row boundary within the reversed (normalized) result
func fetchPageByTimeAndID(ctx context.Context, db *gorm.DB, page keyset.Page, ord keyset.Order) ([]model.Post, string) {
	var out []model.Post

	// Build the query scope (WHERE + ORDER + LIMIT)
	scope := kgorm.PageByTimeAndID(
		db.WithContext(ctx).Model(&model.Post{}),
		page,
		ord,
		"created_at",
		"id",
	)

	// Execute + normalize order (DirPrev → reverse slice)
	if tx := kgorm.FindPage(scope, page, &out); tx.Error != nil {
		log.Fatalf("find page: %v", tx.Error)
	}

	// If there are no rows, no cursor to return
	if len(out) == 0 {
		return out, ""
	}

	// Compute the next cursor from the boundary row in the SAME direction.
	// Because FindPage() normalized order (i.e., output is always display order),
	// the boundary for the "next call" is the last row in 'out'.
	last := out[len(out)-1]
	nextCursor := keyset.EncodeTimeAndInt64Cursor(last.CreatedAt, last.ID)
	return out, nextCursor
}

func printPage(i int, dir keyset.Dir, posts []model.Post) {
	label := "NEXT"
	if dir == keyset.DirPrev {
		label = "PREV"
	}
	fmt.Printf("\n-- %s PAGE %d (len=%d) --\n", label, i, len(posts))
	for _, p := range posts {
		fmt.Printf("ID=%d  CreatedAt=%s  Title=%q\n", p.ID, p.CreatedAt.Format(time.RFC3339), p.Title)
	}
}
