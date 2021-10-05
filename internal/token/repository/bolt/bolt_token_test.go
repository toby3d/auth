//nolint: wrapcheck
package bolt_test

import (
	"context"
	"testing"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
	repository "source.toby3d.me/website/oauth/internal/token/repository/bolt"
	"source.toby3d.me/website/oauth/internal/util"
)

func TestGet(t *testing.T) {
	t.Parallel()

	db, cleanup := util.TestBolt(t, repository.Token{}.Bucket())
	t.Cleanup(cleanup)

	accessToken := domain.TestToken(t)
	accessToken.Profile = nil

	dto := new(repository.Token)
	dto.Populate(accessToken)

	src, err := json.Marshal(dto)
	require.NoError(t, err)

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		return tx.Bucket(repository.Token{}.Bucket()).Put([]byte(accessToken.AccessToken), src)
	}))

	result, err := repository.NewBoltTokenRepository(db).Get(context.TODO(), accessToken.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, accessToken, result)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	db, cleanup := util.TestBolt(t, repository.Token{}.Bucket())
	t.Cleanup(cleanup)

	repo := repository.NewBoltTokenRepository(db)
	accessToken := domain.TestToken(t)
	accessToken.Profile = nil

	require.NoError(t, repo.Create(context.TODO(), accessToken))

	result := new(domain.Token)

	require.NoError(t, db.View(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		return new(repository.Token).Bind(tx.Bucket(repository.Token{}.Bucket()).
			Get([]byte(accessToken.AccessToken)), result)
	}))

	assert.Equal(t, accessToken, result)
	assert.EqualError(t, repo.Create(context.TODO(), accessToken), token.ErrExist.Error())
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	db, cleanup := util.TestBolt(t, repository.Token{}.Bucket())
	t.Cleanup(cleanup)

	accessToken := domain.TestToken(t)

	src, err := json.Marshal(accessToken)
	require.NoError(t, err)

	//nolint: exhaustivestruct
	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(repository.Token{}.Bucket()).Put([]byte(accessToken.AccessToken), src)
	}))

	require.NoError(t, repository.NewBoltTokenRepository(db).Update(context.TODO(), &domain.Token{
		AccessToken: accessToken.AccessToken,
		ClientID:    "https://client.example.net/",
		Me:          "https://toby3d.ru/",
		Scopes:      []string{"read"},
		Type:        "Bearer",
		Profile:     nil,
	}))

	result := domain.NewToken()

	//nolint: exhaustivestruct
	require.NoError(t, db.View(func(tx *bolt.Tx) error {
		return new(repository.Token).Bind(tx.Bucket(repository.Token{}.Bucket()).
			Get([]byte(accessToken.AccessToken)), result)
	}))

	assert.Equal(t, &domain.Token{
		AccessToken: accessToken.AccessToken,
		ClientID:    "https://client.example.net/",
		Me:          "https://toby3d.ru/",
		Scopes:      []string{"read"},
		Type:        "Bearer",
		Profile:     nil,
	}, result)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	db, cleanup := util.TestBolt(t, repository.Token{}.Bucket())
	t.Cleanup(cleanup)

	accessToken := domain.TestToken(t)

	src, err := json.Marshal(accessToken)
	require.NoError(t, err)

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		return tx.Bucket(repository.Token{}.Bucket()).Put([]byte(accessToken.AccessToken), src)
	}))

	require.NoError(t, repository.NewBoltTokenRepository(db).Remove(context.TODO(), accessToken.AccessToken))

	require.NoError(t, db.View(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		assert.Nil(t, tx.Bucket(repository.Token{}.Bucket()).Get([]byte(accessToken.AccessToken)))

		return nil
	}))
}
