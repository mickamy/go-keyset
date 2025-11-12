package keyset

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrCursorLength is returned when a decoded cursor has an unexpected byte length.
	ErrCursorLength = errors.New("invalid cursor length")
)

// EncodeInt64Cursor encodes a signed 64-bit integer into an opaque base64url cursor.
// It uses 8-byte big-endian representation to preserve lexicographical order.
func EncodeInt64Cursor(v int64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(v))
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

// DecodeInt64Cursor decodes a base64url cursor into a signed 64-bit integer.
func DecodeInt64Cursor(s string) (int64, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return 0, fmt.Errorf("keyset: decode cursor: %w", err)
	}
	if len(b) != 8 {
		return 0, ErrCursorLength
	}
	u := binary.BigEndian.Uint64(b)
	return int64(u), nil
}

// EncodeTimeCursor encodes a time value (UTC) as nanoseconds since Unix epoch
// into an opaque base64url cursor. It uses 8-byte big-endian representation
// so lexical order matches chronological order.
func EncodeTimeCursor(t time.Time) string {
	// Strip monotonic component and normalize to UTC before encoding.
	u := uint64(t.UTC().UnixNano())
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], u)
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

// DecodeTimeCursor decodes a time cursor produced by EncodeTimeCursor.
func DecodeTimeCursor(s string) (time.Time, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("keyset: decode cursor: %w", err)
	}
	if len(b) != 8 {
		return time.Time{}, ErrCursorLength
	}
	u := binary.BigEndian.Uint64(b)
	return time.Unix(0, int64(u)).UTC(), nil
}

// EncodeTimeAndInt64Cursor encodes a composite key (time, id) into an opaque
// cursor: 8 bytes for time (ns UTC) + 8 bytes for id, both big-endian.
// This preserves lexicographic order for (time DESC/ASC, id DESC/ASC) when
// appropriate comparison operators are used in SQL.
func EncodeTimeAndInt64Cursor(t time.Time, id int64) string {
	var buf [16]byte
	binary.BigEndian.PutUint64(buf[0:8], uint64(t.UTC().UnixNano()))
	binary.BigEndian.PutUint64(buf[8:16], uint64(id))
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

// DecodeTimeAndInt64Cursor decodes a composite cursor into (time, id).
func DecodeTimeAndInt64Cursor(s string) (time.Time, int64, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("keyset: decode cursor: %w", err)
	}
	if len(b) != 16 {
		return time.Time{}, 0, ErrCursorLength
	}
	tu := binary.BigEndian.Uint64(b[0:8])
	idu := binary.BigEndian.Uint64(b[8:16])
	return time.Unix(0, int64(tu)).UTC(), int64(idu), nil
}
