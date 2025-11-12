# go-keyset

Keyset pagination toolkit for Go.
A minimal and type-safe foundation for stable, bidirectional pagination.

---

## Overview

Offset-based pagination (`OFFSET n LIMIT m`) is simple but inefficient:

* Large offsets require scanning and sorting many rows.
* Insertions or deletions can cause inconsistent page boundaries.

**Keyset pagination** avoids these issues by using stable sort keys (e.g., `id`, `created_at`) and cursor-based
windows (`WHERE created_at < ?`).

This library provides:

* Keyset window builders (`PageByID`, `PageByTime`, `PageByTimeAndID`)
* Direction-aware pagination (`DirNext` / `DirPrev`)
* Opaque cursor encoding (`int64`, `time`, or composite `(time,id)`)
* Utilities for order handling and slice normalization

---

## Packages

```
go-keyset/
  keyset/        # Core logic (Page, Order, cursor encoding)
  kgorm/         # GORM adapter with composable scopes
  examples/      # Practical PostgreSQL example
```

---

## Quick Start

### Installation

```bash
go get github.com/mickamy/go-keyset
go get github.com/mickamy/go-keyset/kgorm # for GORM integration
```

---

### Example (GORM)

```go
import (
    "github.com/mickamy/go-keyset"
    "github.com/mickamy/go-keyset/kgorm"
)

func main() {
    var page = keyset.Page{
        Limit: 5,
        Dir:   keyset.DirNext,
    }
    
    var posts []Post
    
    tx := kgorm.FindPage(
        kgorm.PageByTimeAndID(
            db.WithContext(ctx).Model(&Post{}),
            page,
            keyset.Descending,
            "created_at",
            "id",
        ),
        page,
        &posts,
    )
}
```

**Key points:**

* `PageByTimeAndID` builds a stable keyset window.
* `FindPage` executes the query and automatically normalizes results when using `DirPrev`.

---

## Cursor Encoding

| Type        | Encode                           | Decode                        | Notes                     |
|-------------|----------------------------------|-------------------------------|---------------------------|
| `int64`     | `EncodeInt64Cursor(v)`           | `DecodeInt64Cursor(s)`        | 8-byte big-endian         |
| `time.Time` | `EncodeTimeCursor(t)`            | `DecodeTimeCursor(s)`         | UTC, nanoseconds          |
| `(time,id)` | `EncodeTimeAndInt64Cursor(t,id)` | `DecodeTimeAndInt64Cursor(s)` | Composite for stable sort |

Cursors are opaque base64url strings that are safe for use in URLs and JSON.

---

## License

[MIT](./LICENSE)
