package bolt_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
	repository "source.toby3d.me/website/oauth/internal/token/repository/bolt"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	db, cleanup := util.TestBolt(t, repository.DefaultBucket)
	t.Cleanup(cleanup)

	repo := repository.NewBoltTokenRepository(db)
	accessToken := domain.TestToken(t)

	require.NoError(t, repo.Create(context.TODO(), accessToken))

	result := &domain.Token{
		AccessToken: accessToken.AccessToken,
		TokenType:   accessToken.TokenType,
		Expiry:      time.Time{},
	}

	require.NoError(t, db.View(func(tx *bolt.Tx) (err error) {
		//nolint: exhaustivestruct
		src := tx.Bucket(repository.DefaultBucket).Get([]byte(accessToken.AccessToken))

		result.Expiry, err = time.Parse(time.RFC3339, string(src))

		return
	}))
	assert.Equal(t, accessToken, result)

	assert.ErrorIs(t, repo.Create(context.TODO(), accessToken), token.ErrExist)
}

func TestGet(t *testing.T) {
	t.Parallel()

	db, cleanup := util.TestBolt(t, repository.DefaultBucket)
	t.Cleanup(cleanup)

	repo := repository.NewBoltTokenRepository(db)
	accessToken := domain.TestToken(t)

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(repository.DefaultBucket)
		if err != nil {
			return err
		}

		return bkt.Put([]byte(accessToken.AccessToken), []byte(accessToken.Expiry.Format(time.RFC3339)))
	}))

	result, err := repo.Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken, result)
}
