package bolt

import (
	"context"

	json "github.com/goccy/go-json"
	"gitlab.com/toby3d/indieauth/internal/auth"
	"gitlab.com/toby3d/indieauth/internal/model"
	bolt "go.etcd.io/bbolt"
)

type boltAuthRepository struct {
	db *bolt.DB
}

func NewBoltAuthRepository(db *bolt.DB) (auth.Repository, error) {
	if err := db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(model.Login{}.Bucket())

		return err
	}); err != nil {
		return nil, err
	}

	return &boltAuthRepository{
		db: db,
	}, nil
}

func (repo *boltAuthRepository) Create(ctx context.Context, login *model.Login) error {
	jsonLogin, err := json.Marshal(login)
	if err != nil {
		return err
	}

	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(model.Login{}.Bucket()).Put([]byte(login.Code), jsonLogin)
	})
}

func (repo *boltAuthRepository) Get(ctx context.Context, code string) (*model.Login, error) {
	login := new(model.Login)

	if err := repo.db.View(func(tx *bolt.Tx) error {
		return json.Unmarshal(tx.Bucket(model.Login{}.Bucket()).Get([]byte(code)), login)
	}); err != nil {
		return nil, err
	}

	return login, nil
}

func (repo *boltAuthRepository) Delete(ctx context.Context, code string) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(model.Login{}.Bucket()).Delete([]byte(code))
	})
}
