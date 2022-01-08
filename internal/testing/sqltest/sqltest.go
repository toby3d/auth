package sqltest

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// Open creates a new InMemory sqlite3 database for testing.
func Open(tb testing.TB) (*sqlx.DB, func()) {
	tb.Helper()

	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(tb, err)

	if !assert.NoError(tb, db.Ping()) {
		_ = db.Close() //nolint: errcheck

		tb.FailNow()
	}

	return db, func() {
		_ = db.Close() //nolint: errcheck
	}
}
