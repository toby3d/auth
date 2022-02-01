package bolttest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

// New returns a temporary empty database bbolt in the temporary directory
// with the cleanup function.
func New(tb testing.TB) (*bolt.DB, func()) {
	tb.Helper()

	f, err := os.CreateTemp(os.TempDir(), "bbolt_*.db")
	require.NoError(tb, err)

	filePath := f.Name()
	assert.NoError(tb, f.Close())

	db, err := bolt.Open(filePath, os.ModePerm, nil)
	require.NoError(tb, err)

	return db, func() {
		db.Close()
		os.Remove(filePath)
	}
}
