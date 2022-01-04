package bolt_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/token"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/bolt"
	"source.toby3d.me/website/indieauth/internal/util"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	//nolint: exhaustivestruct
	db, cleanup := util.TestBolt(t, repository.Token{}.Bucket())
	t.Cleanup(cleanup)

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		_, err := tx.CreateBucketIfNotExists(repository.Token{}.Bucket())

		//nolint: wrapcheck
		return err
	}))

	repo := repository.NewBoltTokenRepository(db)
	accessToken := domain.TestToken(t)

	require.NoError(t, repo.Create(context.TODO(), accessToken))

	result := domain.NewToken()

	require.NoError(t, db.View(func(tx *bolt.Tx) (err error) {
		dto := new(repository.Token)

		//nolint: wrapcheck
		return dto.Bind(tx.Bucket(dto.Bucket()).Get([]byte(accessToken.AccessToken)), result)
	}))
	assert.Equal(t, accessToken, result)

	assert.ErrorIs(t, repo.Create(context.TODO(), accessToken), token.ErrExist)
}

func TestGet(t *testing.T) {
	t.Parallel()

	//nolint: exhaustivestruct
	db, cleanup := util.TestBolt(t, repository.Token{}.Bucket())
	t.Cleanup(cleanup)

	accessToken := domain.TestToken(t)

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		bkt, err := tx.CreateBucketIfNotExists(repository.Token{}.Bucket())
		if err != nil {
			return errors.Wrap(err, "cannot create bucket")
		}

		t := new(repository.Token)
		t.Populate(accessToken)

		src, err := json.Marshal(t)
		if err != nil {
			return errors.Wrap(err, "cannot marshal token data")
		}

		if err = bkt.Put([]byte(t.AccessToken), src); err != nil {
			return errors.Wrap(err, "cannot put token into bucket")
		}

		return nil
	}))

	result, err := repository.NewBoltTokenRepository(db).Get(context.TODO(), accessToken.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, accessToken, result)
}
