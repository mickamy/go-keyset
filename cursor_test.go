package keyset_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/mickamy/go-keyset"
)

func TestEncodeDecodeInt64_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := []int64{
		0,
		1,
		-1,
		42,
		-42,
		int64(^uint64(0) >> 1),    // max int64
		-int64(^uint64(0)>>1) - 1, // min int64
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("v=%d", v), func(t *testing.T) {
			t.Parallel()
			cur := keyset.EncodeInt64Cursor(v)
			got, err := keyset.DecodeInt64Cursor(cur)
			if err != nil {
				t.Fatalf("decode failed for %d: %v", v, err)
			}
			if got != v {
				t.Fatalf("round trip mismatch: want %d, got %d", v, got)
			}
		})
	}
}

func TestDecodeInt64_InvalidBase64(t *testing.T) {
	t.Parallel()
	if _, err := keyset.DecodeInt64Cursor("@@@not_base64@@@"); err == nil {
		t.Fatalf("expected error for invalid base64, got nil")
	}
}

func TestDecodeInt64_InvalidLength(t *testing.T) {
	t.Parallel()
	bad := base64.RawURLEncoding.EncodeToString(make([]byte, 7))
	_, err := keyset.DecodeInt64Cursor(bad)
	if !errors.Is(err, keyset.ErrCursorLength) {
		t.Fatalf("expected ErrCursorLength, got %v", err)
	}
}

func TestEncodeDecodeTime_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := []time.Time{
		time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
		time.Date(1999, 12, 31, 23, 59, 59, 999999999, time.UTC),
		time.Date(2038, 1, 19, 3, 14, 7, 0, time.UTC),
	}
	for i, v := range cases {
		v := v
		t.Run(fmt.Sprintf("case=%d", i), func(t *testing.T) {
			t.Parallel()
			cur := keyset.EncodeTimeCursor(v)
			got, err := keyset.DecodeTimeCursor(cur)
			if err != nil {
				t.Fatalf("decode failed for %v: %v", v, err)
			}
			if !got.Equal(v.UTC()) {
				t.Fatalf("round trip mismatch: want %v, got %v", v.UTC(), got)
			}
		})
	}
}

func TestDecodeTime_InvalidLength(t *testing.T) {
	t.Parallel()
	bad := base64.RawURLEncoding.EncodeToString(make([]byte, 7)) // need 8
	_, err := keyset.DecodeTimeCursor(bad)
	if !errors.Is(err, keyset.ErrCursorLength) {
		t.Fatalf("expected ErrCursorLength, got %v", err)
	}
}

func TestEncodeDecodeTimeAndInt64_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		t  time.Time
		id int64
	}{
		{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 1},
		{time.Date(2025, 11, 12, 0, 0, 0, 123456789, time.UTC), 1234567890},
		{time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), -42},
	}
	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("case=%d", i), func(t *testing.T) {
			t.Parallel()
			cur := keyset.EncodeTimeAndInt64Cursor(c.t, c.id)
			gt, gid, err := keyset.DecodeTimeAndInt64Cursor(cur)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if !gt.Equal(c.t.UTC()) || gid != c.id {
				t.Fatalf("round trip mismatch: want (%v,%d), got (%v,%d)", c.t.UTC(), c.id, gt, gid)
			}
		})
	}
}

func TestDecodeTimeAndInt64_InvalidLength(t *testing.T) {
	t.Parallel()
	bad := base64.RawURLEncoding.EncodeToString(make([]byte, 15)) // need 16
	_, _, err := keyset.DecodeTimeAndInt64Cursor(bad)
	if !errors.Is(err, keyset.ErrCursorLength) {
		t.Fatalf("expected ErrCursorLength, got %v", err)
	}
}
