package model

import (
	"time"
)

type Post struct {
	ID        int64
	Title     string
	CreatedAt time.Time
}
