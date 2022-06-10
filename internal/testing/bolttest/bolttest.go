package bolttest

import (
	"os"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// New returns a temporary empty database bbolt in the temporary directory
// with the cleanup function.
func New(tb testing.TB) (*bolt.DB, func()) {
	tb.Helper()

	tempFile, err := os.CreateTemp(os.TempDir(), "bbolt_*.db")
	if err != nil {
		tb.Fatal(err)
	}

	filePath := tempFile.Name()

	if err := tempFile.Close(); err != nil {
		tb.Fatal(err)
	}

	db, err := bolt.Open(filePath, os.ModePerm, nil)
	if err != nil {
		tb.Fatal(err)
	}

	return db, func() {
		_ = db.Close()
		_ = os.Remove(filePath)
	}
}
