// kgorm/pagination_test.go
package kgorm_test

import (
	"strings"
	"testing"
	"time"

	"github.com/mickamy/go-keyset"
	"github.com/mickamy/go-keyset/kgorm"
)

func TestPageByTimeAndID_NoCursor_OrderOnly(t *testing.T) {
	t.Parallel()
	db := openDryRun(t)

	t.Run("ASC DirNext", func(t *testing.T) {
		t.Parallel()
		page := keyset.Page{Limit: 0, Dir: keyset.DirNext}
		sql, vars := toSQL[Post](kgorm.PageByTimeAndID(
			db.Model(&Post{}), page, keyset.Ascending, "created_at", "id",
		))
		if !strings.Contains(sql, "ORDER BY created_at ASC, id ASC") {
			t.Fatalf("missing ORDER ASC, got: %s", sql)
		}
		if !strings.Contains(sql, "LIMIT $") {
			t.Fatalf("missing LIMIT placeholder, got: %s", sql)
		}
		if len(vars) != 1 || vars[0] != 50 {
			t.Fatalf("limit var mismatch: vars=%v", vars)
		}
	})

	t.Run("DESC DirNext", func(t *testing.T) {
		t.Parallel()
		page := keyset.Page{Limit: 0, Dir: keyset.DirNext}
		sql, vars := toSQL[Post](kgorm.PageByTimeAndID(
			db.Model(&Post{}), page, keyset.Descending, "created_at", "id",
		))
		if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC") {
			t.Fatalf("missing ORDER DESC, got: %s", sql)
		}
		if !strings.Contains(sql, "LIMIT $") {
			t.Fatalf("missing LIMIT placeholder, got: %s", sql)
		}
		if len(vars) != 1 || vars[0] != 50 {
			t.Fatalf("limit var mismatch: vars=%v", vars)
		}
	})
}

func TestPageByTimeAndID_WithValidCursor_EffectiveOrderAndWhere(t *testing.T) {
	t.Parallel()
	db := openDryRun(t)

	// DirPrev with base order DESC → effective becomes ASC.
	ts := time.Date(2025, 11, 12, 0, 0, 0, 0, time.UTC)
	cur := keyset.EncodeTimeAndInt64Cursor(ts, 123)
	page := keyset.Page{Cursor: cur, Limit: 10, Dir: keyset.DirPrev}

	sql, vars := toSQL[Post](kgorm.PageByTimeAndID(
		db.Model(&Post{}), page, keyset.Descending, "created_at", "id",
	))

	// WHERE should be ASC-stable window:
	// (created_at > $1) OR (created_at = $2 AND id > $3)
	if !strings.Contains(sql, "(created_at > $1) OR (created_at = $2 AND id > $3)") {
		t.Fatalf("missing ASC stable WHERE, got: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY created_at ASC, id ASC") {
		t.Fatalf("missing ORDER ASC for DirPrev, got: %s", sql)
	}
	if !strings.Contains(sql, "LIMIT $") {
		t.Fatalf("missing LIMIT placeholder, got: %s", sql)
	}
	if len(vars) != 4 {
		t.Fatalf("vars length = %d, want 4 (t,t,id,limit)", len(vars))
	}
	if tm, ok := vars[0].(time.Time); !ok || !tm.Equal(ts) {
		t.Fatalf("vars[0] want time=%v, got %T %v", ts, vars[0], vars[0])
	}
	if tm, ok := vars[1].(time.Time); !ok || !tm.Equal(ts) {
		t.Fatalf("vars[1] want time=%v, got %T %v", ts, vars[1], vars[1])
	}
	if id, ok := vars[2].(int64); !ok || id != 123 {
		t.Fatalf("vars[2] want id=123, got %T %v", vars[2], vars[2])
	}
	if vars[3] != 10 {
		t.Fatalf("vars[3] want limit=10, got %v", vars[3])
	}
}

func TestPageByTimeAndID_InvalidCursor_FallbackNoWhere(t *testing.T) {
	t.Parallel()
	db := openDryRun(t)

	page := keyset.Page{Cursor: "@@@invalid@@@", Limit: 5, Dir: keyset.DirPrev}
	sql, vars := toSQL[Post](kgorm.PageByTimeAndID(
		db.Model(&Post{}), page, keyset.Descending, "created_at", "id",
	))

	// Should fall back to no WHERE clause (we only assert ORDER/LIMIT).
	if strings.Contains(strings.ToUpper(sql), " WHERE ") {
		t.Fatalf("unexpected WHERE with invalid cursor, got: %s", sql)
	}
	// DirPrev with base DESC → ORDER ASC
	if !strings.Contains(sql, "ORDER BY created_at ASC, id ASC") {
		t.Fatalf("missing ORDER ASC on fallback, got: %s", sql)
	}
	if !strings.Contains(sql, "LIMIT $") {
		t.Fatalf("missing LIMIT placeholder, got: %s", sql)
	}
	if len(vars) != 1 || vars[0] != 5 {
		t.Fatalf("limit var mismatch: vars=%v", vars)
	}
}

func TestPageByID_Simple(t *testing.T) {
	t.Parallel()
	db := openDryRun(t)

	t.Run("DESC DirNext with cursor", func(t *testing.T) {
		t.Parallel()
		cur := keyset.EncodeInt64Cursor(100)
		page := keyset.Page{Cursor: cur, Limit: 7, Dir: keyset.DirNext}
		sql, vars := toSQL[Post](kgorm.PageByID(
			db.Model(&Post{}), page, keyset.Descending, "id",
		))
		// WHERE id < $1
		if !strings.Contains(sql, "WHERE id < $1") {
			t.Fatalf("missing WHERE id < $1, got: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY id DESC") {
			t.Fatalf("missing ORDER BY id DESC, got: %s", sql)
		}
		if !strings.Contains(sql, "LIMIT $") {
			t.Fatalf("missing LIMIT placeholder, got: %s", sql)
		}
		if len(vars) != 2 || vars[0] != int64(100) || vars[1] != 7 {
			t.Fatalf("vars mismatch: want [100 7], got %v", vars)
		}
	})

	t.Run("ASC DirPrev no cursor", func(t *testing.T) {
		t.Parallel()
		page := keyset.Page{Limit: 3, Dir: keyset.DirPrev}
		sql, vars := toSQL[Post](kgorm.PageByID(
			db.Model(&Post{}), page, keyset.Ascending, "id",
		))
		// DirPrev with base ASC → ORDER DESC for window.
		if !strings.Contains(sql, "ORDER BY id DESC") {
			t.Fatalf("missing ORDER BY id DESC for DirPrev, got: %s", sql)
		}
		if !strings.Contains(sql, "LIMIT $") {
			t.Fatalf("missing LIMIT placeholder, got: %s", sql)
		}
		if len(vars) != 1 || vars[0] != 3 {
			t.Fatalf("limit var mismatch: vars=%v", vars)
		}
	})
}

func TestPageByTime_Simple(t *testing.T) {
	t.Parallel()
	db := openDryRun(t)

	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	cur := keyset.EncodeTimeCursor(now)

	t.Run("ASC DirNext with cursor", func(t *testing.T) {
		t.Parallel()
		page := keyset.Page{Cursor: cur, Limit: 9, Dir: keyset.DirNext}
		sql, vars := toSQL[Post](kgorm.PageByTime(
			db.Model(&Post{}), page, keyset.Ascending, "created_at",
		))
		// WHERE created_at > $1
		if !strings.Contains(sql, "WHERE created_at > $1") {
			t.Fatalf("missing WHERE created_at > $1, got: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY created_at ASC") {
			t.Fatalf("missing ORDER BY created_at ASC, got: %s", sql)
		}
		if !strings.Contains(sql, "LIMIT $") {
			t.Fatalf("missing LIMIT placeholder, got: %s", sql)
		}
		if len(vars) != 2 || vars[0] != now || vars[1] != 9 {
			t.Fatalf("vars mismatch: want [time, 9], got %v", vars)
		}
	})

	t.Run("DESC DirPrev no cursor", func(t *testing.T) {
		t.Parallel()
		page := keyset.Page{Limit: 0, Dir: keyset.DirPrev}
		sql, vars := toSQL[Post](kgorm.PageByTime(
			db.Model(&Post{}), page, keyset.Descending, "created_at",
		))
		// DirPrev with base DESC → ORDER ASC
		if !strings.Contains(sql, "ORDER BY created_at ASC") {
			t.Fatalf("missing ORDER BY created_at ASC for DirPrev, got: %s", sql)
		}
		if !strings.Contains(sql, "LIMIT $") {
			t.Fatalf("missing LIMIT placeholder, got: %s", sql)
		}
		if len(vars) != 1 || vars[0] != 50 {
			t.Fatalf("limit var mismatch (default 50), vars=%v", vars)
		}
	})
}
