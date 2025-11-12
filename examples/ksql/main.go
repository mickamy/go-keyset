// examples/ksql/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mickamy/go-keyset"
	"github.com/mickamy/go-keyset/examples/internal/model"
	"github.com/mickamy/go-keyset/ksql"
)

const dsn = "postgres://postgres:password@localhost:5432/keyset?sslmode=disable"

func main() {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("== Keyset pagination (created_at DESC, id DESC) ==")

	// Page forward (DirNext) twice.
	var (
		cursor       string
		lastNextPage []model.Post
	)
	for i := 1; i <= 2; i++ {
		page := keyset.Page{Limit: 5, Dir: keyset.DirNext, Cursor: cursor}

		posts, next, err := fetchPosts(ctx, db, page)
		if err != nil {
			log.Fatalf("fetch posts: %v", err)
		}
		fmt.Printf("\n-- NEXT PAGE %d (len=%d) --\n", i, len(posts))
		printPosts(posts)

		cursor = next
		lastNextPage = posts
	}

	// Page backward (DirPrev) from the FIRST item of the last displayed NEXT page.
	// For Prev, the boundary cursor must be derived from the first visible record.
	fmt.Println()
	if len(lastNextPage) == 0 {
		fmt.Println("-- PREV PAGE 1 (len=0) --")
		fmt.Println("(no rows)")
		return
	}
	first := lastNextPage[0]
	prevCursor := keyset.EncodeTimeAndInt64Cursor(first.CreatedAt, first.ID)

	prevPage := keyset.Page{Limit: 5, Dir: keyset.DirPrev, Cursor: prevCursor}
	postsPrev, _, err := fetchPosts(ctx, db, prevPage)
	if err != nil {
		log.Fatalf("fetch prev: %v", err)
	}
	fmt.Printf("-- PREV PAGE 1 (len=%d) --\n", len(postsPrev))
	printPosts(postsPrev)
}

// fetchPosts issues a keyset-paginated query using QueryByTimeAndID with
// (created_at, id) composite key. It returns the rows in DISPLAY ORDER and
// a "next page" cursor derived from the last visible record.
//
// ORDER strategy in this example:
//   - Base order: Descending (newest first)
//   - Next page: pass the last visible boundary (EncodeNextCursor)
//   - Prev page: use DirPrev and pass a cursor derived from the FIRST visible record
func fetchPosts(ctx context.Context, db *sql.DB, page keyset.Page) ([]model.Post, string, error) {
	base := `SELECT id, title, created_at FROM posts`

	// Build SQL and args for the given page/order.
	sqlStr, args := ksql.QueryByTimeAndID(base, page, keyset.Descending, "created_at", "id", ksql.PlaceholderDollar)

	rows, err := db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, "", fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var items []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("scan: %w", err)
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("rows: %w", err)
	}

	// Normalize for display if DirPrev (because SQL ORDER is reversed in that case).
	keyset.NormalizePageResult(page, items)

	// Compute next cursor from the last visible item (display boundary).
	var next string
	if n := len(items); n > 0 {
		last := items[n-1]
		next = keyset.EncodeNextCursor(last.CreatedAt, last.ID)
	}
	return items, next, nil
}

func printPosts(items []model.Post) {
	if len(items) == 0 {
		fmt.Println("(no rows)")
		return
	}
	for _, p := range items {
		fmt.Printf("id=%d  created_at=%s  title=%s\n", p.ID, p.CreatedAt.Format(time.RFC3339), p.Title)
	}
}
