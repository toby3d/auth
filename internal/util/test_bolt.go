package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestBolt(tb testing.TB, buckets ...[]byte) (*bolt.DB, func()) {
	tb.Helper()

	f, err := os.CreateTemp("", "bbolt_*.db")
	require.NoError(tb, err)

	filePath := f.Name()
	assert.NoError(tb, f.Close())

	db, err := bolt.Open(filePath, os.ModePerm, nil)
	require.NoError(tb, err)

	for _, bucket := range buckets {
		bucket := bucket

		assert.NoError(tb, db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket(bucket)

			return err //nolint: errcheck
		}))
	}

	return db, func() {
		db.Close()          //nolint: errcheck
		os.Remove(filePath) //nolint: errcheck
	}
}
