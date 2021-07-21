package bolt

import (
	"context"

	json "github.com/goccy/go-json"
	"gitlab.com/toby3d/indieauth/internal/model"
	"gitlab.com/toby3d/indieauth/internal/token"
	bolt "go.etcd.io/bbolt"
)

type boltTokenRepository struct {
	db *bolt.DB
}

func NewBoltTokenRepository(db *bolt.DB) (token.Repository, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(model.Token{}.Bucket())

		return err
	}); err != nil {
		return nil, err
	}

	return &boltTokenRepository{
		db: db,
	}, nil
}

func (repo *boltTokenRepository) Create(ctx context.Context, token *model.Token) error {
	jsonToken, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(model.Token{}.Bucket()).Put([]byte(token.AccessToken), jsonToken)
	})
}

func (repo *boltTokenRepository) Get(ctx context.Context, token string) (*model.Token, error) {
	t := new(model.Token)

	if err := repo.db.View(func(tx *bolt.Tx) error {
		return json.Unmarshal(tx.Bucket(model.Token{}.Bucket()).Get([]byte(token)), t)
	}); err != nil {
		return nil, err
	}

	return t, nil
}

func (repo *boltTokenRepository) Delete(ctx context.Context, token string) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(model.Token{}.Bucket()).Delete([]byte(token))
	})
}
