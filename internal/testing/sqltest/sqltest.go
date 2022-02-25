package sqltest

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Time struct{}

func (Time) Match(v driver.Value) bool {
	_, ok := v.(time.Time)

	return ok
}

// Open creates a new InMemory sqlite3 database for testing.
func Open(tb testing.TB) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	tb.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		tb.Fatal(err)
	}

	xdb := sqlx.NewDb(db, "sqlite3")
	if err = xdb.Ping(); err != nil {
		_ = db.Close()

		tb.Fatal(err)
	}

	return xdb, mock, func() {
		_ = db.Close()
	}
}
