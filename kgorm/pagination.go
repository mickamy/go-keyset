package kgorm

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/mickamy/go-keyset"
)

// PageByID applies keyset pagination over a single integer key column.
//
// Behavior:
//   - Uses opaque cursors produced by keyset.EncodeInt64Cursor.
//   - For DirPrev, it reverses ORDER BY to fetch the previous window.
//   - Use FindPage (or keyset.NormalizePageResult) to restore display order for DirPrev.
func PageByID(db *gorm.DB, p keyset.Page, ord keyset.Order, col string) *gorm.DB {
	p.EnsureDefaults()
	effective := keyset.EffectiveOrder(ord, p.Dir)

	// Decode cursor; if invalid, log a warning and fall back to no WHERE.
	if p.Cursor != "" {
		id, err := keyset.DecodeInt64Cursor(p.Cursor)
		if err != nil {
			db.Logger.Warn(db.Statement.Context, "invalid pagination cursor: cursor=%v error=%v", p.Cursor, err)
		} else {
			// Apply the inequality operator based on effective order.
			db = db.Where(fmt.Sprintf("%s %s ?", col, effective.InequalityOp()), id)
		}
	}

	// Apply ORDER BY and LIMIT.
	return db.Order(fmt.Sprintf("%s %s", col, effective.SQLKeyword())).Limit(p.Limit)
}

// PageByTime applies keyset pagination over a single time column.
// The cursor must be produced by keyset.EncodeTimeCursor.
//
// Note: Use FindPage (or keyset.NormalizePageResult) to restore display order for DirPrev.
func PageByTime(db *gorm.DB, p keyset.Page, ord keyset.Order, col string) *gorm.DB {
	p.EnsureDefaults()
	effective := keyset.EffectiveOrder(ord, p.Dir)

	if p.Cursor != "" {
		tm, err := keyset.DecodeTimeCursor(p.Cursor)
		if err != nil {
			db.Logger.Warn(db.Statement.Context, "invalid pagination cursor: cursor=%v error=%v", p.Cursor, err)
		} else {
			// Apply the inequality operator based on effective order.
			db = db.Where(fmt.Sprintf("%s %s ?", col, effective.InequalityOp()), tm)
		}
	}

	// Apply ORDER BY and LIMIT.
	return db.Order(fmt.Sprintf("%s %s", col, effective.SQLKeyword())).Limit(p.Limit)
}

// PageByTimeAndID applies keyset pagination over a composite key (time, id).
// Columns must refer to the time and id columns respectively.
// Cursor must be produced by keyset.EncodeTimeAndInt64Cursor.
//
// Stable window (DESC):
//
//	(time < :t) OR (time = :t AND id < :id)
//
// Stable window (ASC):
//
//	(time > :t) OR (time = :t AND id > :id)
//
// Note: Use FindPage (or keyset.NormalizePageResult) to restore display order for DirPrev.
func PageByTimeAndID(db *gorm.DB, p keyset.Page, ord keyset.Order, timeCol, idCol string) *gorm.DB {
	p.EnsureDefaults()
	effective := keyset.EffectiveOrder(ord, p.Dir)

	if p.Cursor != "" {
		tm, id, err := keyset.DecodeTimeAndInt64Cursor(p.Cursor)
		if err != nil {
			db.Logger.Warn(db.Statement.Context, "invalid pagination cursor: cursor=%v error=%v", p.Cursor, err)
		} else {
			// Build the stable composite WHERE fragment and bind values (t, t, id).
			where := keyset.StableWhereTimeAndID(timeCol, idCol, effective)
			db = db.Where(where, tm, tm, id)
		}
	}

	// Apply composite ORDER BY and LIMIT.
	order := keyset.OrderClause([]string{timeCol, idCol}, effective)
	return db.Order(order).Limit(p.Limit)
}
