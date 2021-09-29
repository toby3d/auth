//nolint: wrapcheck
package bolt_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/random"
	"source.toby3d.me/website/oauth/internal/token"
	"source.toby3d.me/website/oauth/internal/token/repository/bolt"
)

//nolint: gochecknoglobals
var (
	db   *bbolt.DB
	repo token.Repository
)

func TestMain(m *testing.M) {
	var err error

	dbPath := filepath.Join("..", "..", "..", "..", "test", "testing.db")
	if db, err = bbolt.Open(dbPath, os.ModePerm, nil); err != nil {
		log.Fatalln(err)
	}

	if repo, err = bolt.NewBoltTokenRepository(db); err != nil {
		_ = db.Close()

		log.Fatalln(err)
	}

	code := m.Run()
	_ = db.Close()
	_ = os.RemoveAll(dbPath)

	os.Exit(code)
}

func TestGet(t *testing.T) {
	t.Parallel()

	accessToken := random.New().String(32)

	t.Cleanup(func() {
		_ = db.Update(func(tx *bbolt.Tx) error {
			//nolint: exhaustivestruct
			return tx.Bucket(bolt.Token{}.Bucket()).Delete([]byte(accessToken))
		})
	})

	src, err := json.Marshal(&bolt.Token{
		AccessToken: accessToken,
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Scope:       "read update delete",
		Type:        "Bearer",
	})
	require.NoError(t, err)

	require.NoError(t, db.Update(func(tx *bbolt.Tx) error {
		//nolint: exhaustivestruct
		return tx.Bucket(bolt.Token{}.Bucket()).Put([]byte(accessToken), src)
	}))

	tkn, err := repo.Get(context.TODO(), accessToken)
	require.NoError(t, err)
	assert.Equal(t, &domain.Token{
		AccessToken: accessToken,
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Scopes:      []string{"read", "update", "delete"},
		Type:        "Bearer",
		Profile:     nil,
	}, tkn)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	accessToken := random.New().String(32)

	t.Cleanup(func() {
		_ = db.Update(func(tx *bbolt.Tx) error {
			//nolint: exhaustivestruct
			return tx.Bucket(bolt.Token{}.Bucket()).Delete([]byte(accessToken))
		})
	})

	tkn := &domain.Token{
		AccessToken: accessToken,
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Scopes:      []string{"read", "update", "delete"},
		Type:        "Bearer",
		Profile:     nil,
	}

	require.NoError(t, repo.Create(context.TODO(), tkn))

	result := domain.NewToken()

	require.NoError(t, db.View(func(tx *bbolt.Tx) error {
		//nolint: exhaustivestruct
		return new(bolt.Token).Bind(tx.Bucket(bolt.Token{}.Bucket()).Get([]byte(tkn.AccessToken)), result)
	}))
	assert.Equal(t, tkn, result)

	assert.EqualError(t, repo.Create(context.TODO(), tkn), token.ErrExist.Error())
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	accessToken := random.New().String(32)

	t.Cleanup(func() {
		_ = db.Update(func(tx *bbolt.Tx) error {
			//nolint: exhaustivestruct
			return tx.Bucket(bolt.Token{}.Bucket()).Delete([]byte(accessToken))
		})
	})

	src, err := json.Marshal(&bolt.Token{
		AccessToken: accessToken,
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Scope:       "read update delete",
		Type:        "Bearer",
	})
	require.NoError(t, err)

	require.NoError(t, db.Update(func(tx *bbolt.Tx) error {
		//nolint: exhaustivestruct
		return tx.Bucket(bolt.Token{}.Bucket()).Put([]byte(accessToken), src)
	}))

	require.NoError(t, repo.Update(context.TODO(), &domain.Token{
		AccessToken: accessToken,
		ClientID:    "https://client.example.com/",
		Me:          "https://toby3d.ru/",
		Scopes:      []string{"read"},
		Type:        "Bearer",
		Profile:     nil,
	}))

	result := domain.NewToken()

	require.NoError(t, db.View(func(tx *bbolt.Tx) error {
		//nolint: exhaustivestruct
		return new(bolt.Token).Bind(tx.Bucket(bolt.Token{}.Bucket()).Get([]byte(accessToken)), result)
	}))
	assert.Equal(t, &domain.Token{
		AccessToken: accessToken,
		ClientID:    "https://client.example.com/",
		Me:          "https://toby3d.ru/",
		Scopes:      []string{"read"},
		Type:        "Bearer",
		Profile:     nil,
	}, result)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	accessToken := random.New().String(32)

	t.Cleanup(func() {
		_ = db.Update(func(tx *bbolt.Tx) error {
			//nolint: exhaustivestruct
			return tx.Bucket(bolt.Token{}.Bucket()).Delete([]byte(accessToken))
		})
	})

	src, err := json.Marshal(&bolt.Token{
		AccessToken: accessToken,
		ClientID:    "https://app.example.com/",
		Me:          "https://toby3d.me/",
		Scope:       "read update delete",
		Type:        "Bearer",
	})
	require.NoError(t, err)

	require.NoError(t, db.Update(func(tx *bbolt.Tx) error {
		//nolint: exhaustivestruct
		return tx.Bucket(bolt.Token{}.Bucket()).Put([]byte(accessToken), src)
	}))

	require.NoError(t, repo.Remove(context.TODO(), accessToken))

	require.NoError(t, db.View(func(tx *bbolt.Tx) error {
		//nolint: exhaustivestruct
		assert.Nil(t, tx.Bucket(bolt.Token{}.Bucket()).Get([]byte(accessToken)))

		return nil
	}))
}
