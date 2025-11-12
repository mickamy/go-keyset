package keyset

import (
	"slices"
)

// NormalizePageResult reverses the given result slice in-place
// if the page direction is DirPrev. It returns the same slice for fluency.
func NormalizePageResult[T any](p Page, results []T) []T {
	if p.Dir == DirPrev {
		slices.Reverse(results)
	}
	return results
}
