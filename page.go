package keyset

import (
	"fmt"
)

// Dir represents the direction of pagination.
// DirNext fetches the next page (after the cursor), DirPrev fetches the previous page (before the cursor).
type Dir int

const (
	DirNext Dir = iota + 1
	DirPrev
)

// Order represents sorting order (ascending or descending).
type Order int

const (
	Ascending Order = iota + 1
	Descending
)

// SQLKeyword returns the SQL keyword ("ASC" or "DESC")
// representing the given sort order. It defaults to "ASC".
func (o Order) SQLKeyword() string {
	switch o {
	case Ascending:
		return "ASC"
	case Descending:
		return "DESC"
	default:
		return "ASC"
	}
}

// Reverse returns the opposite sort order.
// Ascending → Descending, Descending → Ascending.
func (o Order) Reverse() Order {
	switch o {
	case Ascending:
		return Descending
	case Descending:
		return Ascending
	default:
		panic(fmt.Sprintf("invalid order: %d", o))
	}
}

// InequalityOp returns the comparison operator (either "<" or ">")
// to be used for single-column keyset windows under this order.
// Descending → "<", Ascending → ">".
func (o Order) InequalityOp() string {
	switch o {
	case Ascending:
		return ">"
	case Descending:
		return "<"
	default:
		panic(fmt.Sprintf("invalid order: %d", o))
	}
}

// Page defines a keyset pagination state.
type Page struct {
	Cursor string // Encoded opaque cursor string
	Limit  int    // Number of items to fetch
	Dir    Dir    // Direction of pagination
}

// EnsureDefaults fills unset fields with default values.
// It does not report invalid states; use Validate for strict checks.
func (p *Page) EnsureDefaults() {
	if p.Limit <= 0 {
		p.Limit = 50
	}
	if p.Dir != DirNext && p.Dir != DirPrev {
		p.Dir = DirNext
	}
}
