package ksql_test

import (
	"strings"
	"testing"
	"time"

	"github.com/mickamy/go-keyset"
	"github.com/mickamy/go-keyset/ksql"
)

func TestQueryByID_NoCursor_OrderOnly(t *testing.T) {
	t.Parallel()

	base := `SELECT id,title FROM posts`

	t.Run("ASC DirNext default limit", func(t *testing.T) {
		t.Parallel()
		p := keyset.Page{Limit: 0, Dir: keyset.DirNext} // EnsureDefaults → 50
		sql, args := ksql.QueryByID(base, p, keyset.Ascending, "id", ksql.PlaceholderDollar)

		if !strings.Contains(sql, "ORDER BY id ASC") {
			t.Fatalf("missing ORDER ASC: %s", sql)
		}
		if !strings.Contains(sql, " LIMIT $1") {
			t.Fatalf("missing LIMIT placeholder: %s", sql)
		}
		if len(args) != 1 || args[0] != 50 {
			t.Fatalf("limit args mismatch: %v", args)
		}
	})

	t.Run("DESC DirNext default limit", func(t *testing.T) {
		t.Parallel()
		p := keyset.Page{Limit: 0, Dir: keyset.DirNext}
		sql, args := ksql.QueryByID(base, p, keyset.Descending, "id", ksql.PlaceholderDollar)

		if !strings.Contains(sql, "ORDER BY id DESC") {
			t.Fatalf("missing ORDER DESC: %s", sql)
		}
		if !strings.Contains(sql, " LIMIT $1") {
			t.Fatalf("missing LIMIT placeholder: %s", sql)
		}
		if len(args) != 1 || args[0] != 50 {
			t.Fatalf("limit args mismatch: %v", args)
		}
	})
}

func TestQueryByID_WithCursor_And_Prev(t *testing.T) {
	t.Parallel()

	base := `SELECT id FROM posts`

	t.Run("DESC DirNext with cursor", func(t *testing.T) {
		t.Parallel()
		cur := keyset.EncodeInt64Cursor(100)
		p := keyset.Page{Cursor: cur, Limit: 7, Dir: keyset.DirNext}

		sql, args := ksql.QueryByID(base, p, keyset.Descending, "id", ksql.PlaceholderDollar)

		if !strings.Contains(sql, "WHERE id < $1") {
			t.Fatalf("missing WHERE id < $1: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY id DESC") {
			t.Fatalf("missing ORDER BY id DESC: %s", sql)
		}
		if !strings.HasSuffix(sql, " LIMIT $2") {
			t.Fatalf("missing trailing LIMIT $2: %s", sql)
		}
		if len(args) != 2 || args[0] != int64(100) || args[1] != 7 {
			t.Fatalf("args mismatch: %v", args)
		}
	})

	t.Run("ASC DirPrev no cursor (effective DESC)", func(t *testing.T) {
		t.Parallel()
		p := keyset.Page{Limit: 3, Dir: keyset.DirPrev}

		sql, args := ksql.QueryByID(base, p, keyset.Ascending, "id", ksql.PlaceholderDollar)

		if strings.Contains(sql, " WHERE ") {
			t.Fatalf("unexpected WHERE without cursor: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY id DESC") {
			t.Fatalf("DirPrev with base ASC should flip to DESC: %s", sql)
		}
		if !strings.HasSuffix(sql, " LIMIT $1") || len(args) != 1 || args[0] != 3 {
			t.Fatalf("limit args mismatch: %v (sql=%s)", args, sql)
		}
	})
}

func TestQueryByTime_WithAndWithoutCursor(t *testing.T) {
	t.Parallel()

	base := `SELECT created_at FROM posts`
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	cur := keyset.EncodeTimeCursor(now)

	t.Run("ASC DirNext with cursor", func(t *testing.T) {
		t.Parallel()
		p := keyset.Page{Cursor: cur, Limit: 9, Dir: keyset.DirNext}

		sql, args := ksql.QueryByTime(base, p, keyset.Ascending, "created_at", ksql.PlaceholderDollar)

		if !strings.Contains(sql, "WHERE created_at > $1") {
			t.Fatalf("missing WHERE created_at > $1: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY created_at ASC") {
			t.Fatalf("missing ORDER BY created_at ASC: %s", sql)
		}
		if !strings.HasSuffix(sql, " LIMIT $2") {
			t.Fatalf("missing LIMIT $2: %s", sql)
		}
		if len(args) != 2 || args[0] != now || args[1] != 9 {
			t.Fatalf("args mismatch: %v", args)
		}
	})

	t.Run("DESC DirPrev no cursor (effective ASC)", func(t *testing.T) {
		t.Parallel()
		p := keyset.Page{Limit: 0, Dir: keyset.DirPrev}

		sql, args := ksql.QueryByTime(base, p, keyset.Descending, "created_at", ksql.PlaceholderDollar)

		if strings.Contains(sql, " WHERE ") {
			t.Fatalf("unexpected WHERE without cursor: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY created_at ASC") {
			t.Fatalf("DirPrev with base DESC should flip to ASC: %s", sql)
		}
		if !strings.HasSuffix(sql, " LIMIT $1") || len(args) != 1 || args[0] != 50 {
			t.Fatalf("default limit arg mismatch: %v (sql=%s)", args, sql)
		}
	})
}

func TestQueryByTimeAndID_WithCursor_And_InvalidFallback(t *testing.T) {
	t.Parallel()

	base := `SELECT * FROM posts`

	t.Run("DirPrev + base DESC → effective ASC (valid cursor)", func(t *testing.T) {
		t.Parallel()
		ts := time.Date(2025, 11, 12, 0, 0, 0, 0, time.UTC)
		cur := keyset.EncodeTimeAndInt64Cursor(ts, 123)
		p := keyset.Page{Cursor: cur, Limit: 10, Dir: keyset.DirPrev}

		sql, args := ksql.QueryByTimeAndID(base, p, keyset.Descending, "created_at", "id", ksql.PlaceholderDollar)

		if !strings.Contains(sql, "(created_at > $1) OR (created_at = $2 AND id > $3)") {
			t.Fatalf("missing ASC stable WHERE: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY created_at ASC, id ASC") {
			t.Fatalf("missing ORDER ASC,ASC: %s", sql)
		}
		if !strings.HasSuffix(sql, " LIMIT $4") {
			t.Fatalf("missing LIMIT $4: %s", sql)
		}
		if len(args) != 4 {
			t.Fatalf("args length want 4, got %d (%v)", len(args), args)
		}
		if tm, ok := args[0].(time.Time); !ok || !tm.Equal(ts) {
			t.Fatalf("args[0] want time=%v, got %T %v", ts, args[0], args[0])
		}
		if tm, ok := args[1].(time.Time); !ok || !tm.Equal(ts) {
			t.Fatalf("args[1] want time=%v, got %T %v", ts, args[1], args[1])
		}
		if id, ok := args[2].(int64); !ok || id != 123 {
			t.Fatalf("args[2] want id=123, got %T %v", args[2], args[2])
		}
		if args[3] != 10 {
			t.Fatalf("args[3] want limit=10, got %v", args[3])
		}
	})

	t.Run("invalid cursor → no WHERE, preserves existing WHERE via AND", func(t *testing.T) {
		t.Parallel()
		// base already has WHERE; invalid cursor must not add any extra WHERE/AND
		baseWithWhere := `SELECT * FROM posts WHERE status = 'published'`
		p := keyset.Page{Cursor: "@@@invalid@@@", Limit: 5, Dir: keyset.DirNext}

		sql, args := ksql.QueryByTimeAndID(baseWithWhere, p, keyset.Descending, "created_at", "id", ksql.PlaceholderDollar)

		// Should not append a second WHERE/AND (invalid cursor → no window).
		if strings.Contains(sql, " AND (created_at ") || strings.Contains(sql, " WHERE (created_at ") {
			t.Fatalf("unexpected WHERE/AND on invalid cursor: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC") {
			t.Fatalf("missing ORDER DESC, DESC: %s", sql)
		}
		if !strings.HasSuffix(sql, " LIMIT $1") || len(args) != 1 || args[0] != 5 {
			t.Fatalf("limit args mismatch: %v (sql=%s)", args, sql)
		}
	})
}

func TestQueryByTimeAndID_AppendWhereWhenBaseHasWhere(t *testing.T) {
	t.Parallel()

	// Valid cursor + base already has WHERE → append AND (not a new WHERE)
	base := `SELECT * FROM posts WHERE tenant_id = 1`
	ts := time.Unix(0, 0).UTC()
	cur := keyset.EncodeTimeAndInt64Cursor(ts, 9)
	p := keyset.Page{Cursor: cur, Limit: 2, Dir: keyset.DirNext}

	sql, _ := ksql.QueryByTimeAndID(base, p, keyset.Descending, "created_at", "id", ksql.PlaceholderDollar)

	if !strings.Contains(sql, " WHERE tenant_id = 1 AND (created_at < $1) OR (created_at = $2 AND id < $3)") &&
		!strings.Contains(sql, " WHERE tenant_id = 1 AND ((created_at < $1) OR (created_at = $2 AND id < $3))") {
		// Depending on formatting, builders might or might not wrap the OR-group.
		t.Fatalf("expected appended AND with stable window, got: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC") {
		t.Fatalf("missing ORDER DESC, DESC: %s", sql)
	}
}
