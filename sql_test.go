package keyset_test

import (
	"testing"

	"github.com/mickamy/go-keyset"
)

func TestEffectiveOrder(t *testing.T) {
	t.Parallel()

	if keyset.EffectiveOrder(keyset.Ascending, keyset.DirNext) != keyset.Ascending {
		t.Fatalf("DirNext should keep order")
	}
	if keyset.EffectiveOrder(keyset.Ascending, keyset.DirPrev) != keyset.Descending {
		t.Fatalf("DirPrev should reverse order")
	}
}

func TestStableWhereTimeAndID(t *testing.T) {
	t.Parallel()

	desc := keyset.StableWhereTimeAndID("created_at", "id", keyset.Descending)
	if desc != "(created_at < ?) OR (created_at = ? AND id < ?)" {
		t.Fatalf("unexpected DESC where: %s", desc)
	}
	asc := keyset.StableWhereTimeAndID("created_at", "id", keyset.Ascending)
	if asc != "(created_at > ?) OR (created_at = ? AND id > ?)" {
		t.Fatalf("unexpected ASC where: %s", asc)
	}
}

func TestOrderClause(t *testing.T) {
	t.Parallel()

	got := keyset.OrderClause([]string{"created_at", "id"}, keyset.Descending)
	want := "created_at DESC, id DESC"
	if got != want {
		t.Fatalf("unexpected ORDER clause: want %q got %q", want, got)
	}
}
