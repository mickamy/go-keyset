package kgorm_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Post struct {
	ID        int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Title     string
}

// openDryRun returns a GORM *DB configured for DryRun so we can inspect
// generated SQL and bound variables without hitting a real database.
func openDryRun(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	dial := postgres.New(postgres.Config{
		Conn: sqlDB,
	})

	db, err := gorm.Open(dial, &gorm.Config{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("open dryrun db: %v", err)
	}

	return db
}

// toSQL builds and "executes" (DryRun) the query and returns SQL and Vars.
func toSQL[T any](q *gorm.DB) (string, []any) {
	var out []T
	tx := q.Find(&out)
	return tx.Statement.SQL.String(), tx.Statement.Vars
}
