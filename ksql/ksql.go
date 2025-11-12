// Package ksql provides lightweight helpers for building keyset pagination
// queries using database/sql or other low-level SQL interfaces.
//
// It is ORM-agnostic, focusing purely on SQL string composition and cursor logic.
// This allows consistent keyset pagination across different data access layers,
// such as pgx, sqlx, or plain database/sql.
//
// Example:
//
//	base := `SELECT id, title, created_at FROM posts`
//	page := keyset.Page{Limit: 10, Dir: keyset.DirNext, Cursor: ""}
//
//	query, args := ksql.QueryByTimeAndID(
//		base, page, keyset.Descending,
//		"created_at", "id", ksql.PlaceholderDollar,
//	)
//
//	rows, err := db.QueryContext(ctx, query, args...)
//	if err != nil { ... }
//
//	for rows.Next() {
//		// scan rows...
//	}
//
//	// Compute next cursor for HTTP API responses
//	next := keyset.EncodeNextCursor(last.CreatedAt, last.ID)
//
// See also the `examples/ksql` package for a working PostgreSQL example.
package ksql

import (
	"fmt"
	"strings"

	"github.com/mickamy/go-keyset"
)

// Placeholder renders a bind placeholder for the N-th parameter in a query.
// Example for Postgres: "$1", "$2", ...
// Example for MySQL/SQLite: "?" (index is ignored)
type Placeholder func(n int) string

// PlaceholderQuestion returns "?" for any index (MySQL/SQLite style).
func PlaceholderQuestion(_ int) string { return "?" }

// PlaceholderDollar returns "$<n>" (PostgreSQL style).
func PlaceholderDollar(n int) string { return fmt.Sprintf("$%d", n) }

// QueryByID builds a keyset-paginated SQL statement for a single integer key column.
// - base: a SELECT ... FROM ... [WHERE ...] prefix (without ORDER/LIMIT).
// - p:    pagination state (Limit/Dir/Cursor).
// - ord:  base sort order (Ascending/Descending).
// - col:  name of the integer key column.
// - ph:   placeholder strategy (e.g., PlaceholderDollar for Postgres).
//
// The returned SQL appends a stable WHERE window (if a valid cursor is present),
// an ORDER BY clause according to the effective order, and a LIMIT clause.
// The returned args are the bound variables in order (window values followed by limit).
func QueryByID(base string, p keyset.Page, ord keyset.Order, col string, ph Placeholder) (string, []any) {
	p.EnsureDefaults()
	eff := keyset.EffectiveOrder(ord, p.Dir)

	var (
		sqlBuilder strings.Builder
		args       []any
		argIdx     = 1
	)
	sqlBuilder.WriteString(base)

	// WHERE window (if cursor is valid)
	if p.Cursor != "" {
		if id, err := keyset.DecodeInt64Cursor(p.Cursor); err == nil {
			where := fmt.Sprintf("%s %s %s", col, eff.InequalityOp(), ph(argIdx))
			sqlBuilder.WriteString(appendWhere(base, where))
			args = append(args, id)
			argIdx++
		}
		// On invalid cursor: fail open (no WHERE), consistent with kgorm behavior.
	}

	// ORDER BY
	sqlBuilder.WriteString(" ORDER BY ")
	sqlBuilder.WriteString(col)
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(eff.SQLKeyword())

	// LIMIT
	sqlBuilder.WriteString(" LIMIT ")
	sqlBuilder.WriteString(ph(argIdx))
	args = append(args, p.Limit)

	return sqlBuilder.String(), args
}

// QueryByTime builds a keyset-paginated SQL statement for a single time column.
// The cursor must be produced by keyset.EncodeTimeCursor.
// See QueryByID for parameter semantics.
func QueryByTime(base string, p keyset.Page, ord keyset.Order, col string, ph Placeholder) (string, []any) {
	p.EnsureDefaults()
	eff := keyset.EffectiveOrder(ord, p.Dir)

	var (
		sqlBuilder strings.Builder
		args       []any
		argIdx     = 1
	)
	sqlBuilder.WriteString(base)

	if p.Cursor != "" {
		if tm, err := keyset.DecodeTimeCursor(p.Cursor); err == nil {
			where := fmt.Sprintf("%s %s %s", col, eff.InequalityOp(), ph(argIdx))
			sqlBuilder.WriteString(appendWhere(base, where))
			args = append(args, tm)
			argIdx++
		}
	}

	sqlBuilder.WriteString(" ORDER BY ")
	sqlBuilder.WriteString(col)
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(eff.SQLKeyword())

	sqlBuilder.WriteString(" LIMIT ")
	sqlBuilder.WriteString(ph(argIdx))
	args = append(args, p.Limit)

	return sqlBuilder.String(), args
}

// QueryByTimeAndID builds a keyset-paginated SQL statement for the composite key (time, id).
// The cursor must be produced by keyset.EncodeTimeAndInt64Cursor.
// The stable window is:
//
//	DESC: (time < :t) OR (time = :t AND id < :id)
//	ASC : (time > :t) OR (time = :t AND id > :id)
//
// The function appends WHERE (if cursor valid), composite ORDER BY, and LIMIT.
func QueryByTimeAndID(base string, p keyset.Page, ord keyset.Order, timeCol, idCol string, ph Placeholder) (string, []any) {
	p.EnsureDefaults()
	eff := keyset.EffectiveOrder(ord, p.Dir)

	var (
		sqlBuilder strings.Builder
		args       []any
		argIdx     = 1
	)
	sqlBuilder.WriteString(base)

	if p.Cursor != "" {
		if tm, id, err := keyset.DecodeTimeAndInt64Cursor(p.Cursor); err == nil {
			// Build stable WHERE using the effective order.
			tCmp, idCmp := ">", ">"
			if eff == keyset.Descending {
				tCmp, idCmp = "<", "<"
			}
			where := fmt.Sprintf("(%s %s %s) OR (%s = %s AND %s %s %s)",
				timeCol, tCmp, ph(argIdx),
				timeCol, ph(argIdx+1), idCol, idCmp, ph(argIdx+2),
			)
			sqlBuilder.WriteString(appendWhere(base, where))
			args = append(args, tm, tm, id)
			argIdx += 3
		}
	}

	// ORDER BY time, id
	sqlBuilder.WriteString(" ORDER BY ")
	sqlBuilder.WriteString(timeCol)
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(eff.SQLKeyword())
	sqlBuilder.WriteString(", ")
	sqlBuilder.WriteString(idCol)
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(eff.SQLKeyword())

	// LIMIT
	sqlBuilder.WriteString(" LIMIT ")
	sqlBuilder.WriteString(ph(argIdx))
	args = append(args, p.Limit)

	return sqlBuilder.String(), args
}

// appendWhere decides whether to append " WHERE <cond>" or " AND <cond>"
// depending on whether the base already contains a WHERE clause.
// It preserves a leading space.
func appendWhere(base, cond string) string {
	if hasWhere(strings.ToLower(base)) {
		return " AND " + cond
	}
	return " WHERE " + cond
}

func hasWhere(lowerBase string) bool {
	// Naive but effective: search for " where " (surrounded by spaces) or ending with " where"
	// to avoid matching column names or table aliases containing "where".
	if strings.Contains(lowerBase, " where ") || strings.HasSuffix(lowerBase, " where") {
		return true
	}
	return false
}
