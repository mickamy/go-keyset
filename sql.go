package keyset

import "strings"

// EffectiveOrder returns ord if DirNext or the reversed order if DirPrev.
// This is a common utility across adapters to compute the ORDER used for the SQL window.
func EffectiveOrder(ord Order, dir Dir) Order {
	if dir == DirPrev {
		return ord.Reverse()
	}
	return ord
}

// StableWhereTimeAndID builds the stable composite key condition for keyset windows.
//
// For Descending order it returns the SQL fragment:
//
//	(timeCol < ?) OR (timeCol = ? AND idCol < ?)
//
// For Ascending order it returns the SQL fragment:
//
//	(timeCol > ?) OR (timeCol = ? AND idCol > ?)
//
// The placeholders are intended to be bound with (t, t, id) in that order.
// The function only composes the SQL fragment and is ORM-agnostic.
func StableWhereTimeAndID(timeCol, idCol string, ord Order) string {
	var tCmp, idCmp string
	if ord == Descending {
		tCmp, idCmp = "<", "<"
	} else {
		tCmp, idCmp = ">", ">"
	}
	var b strings.Builder
	b.WriteString("(")
	b.WriteString(timeCol)
	b.WriteString(" ")
	b.WriteString(tCmp)
	b.WriteString(" ?)")
	b.WriteString(" OR (")
	b.WriteString(timeCol)
	b.WriteString(" = ? AND ")
	b.WriteString(idCol)
	b.WriteString(" ")
	b.WriteString(idCmp)
	b.WriteString(" ?)")
	return b.String()
}

// OrderClause returns a comma-joined ORDER BY clause for the provided columns
// using the given order. Example: OrderClause([]string{"created_at","id"}, Descending)
// returns: "created_at DESC, id DESC".
func OrderClause(cols []string, ord Order) string {
	kw := ord.SQLKeyword()
	var b strings.Builder
	for i, c := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c)
		b.WriteString(" ")
		b.WriteString(kw)
	}
	return b.String()
}
