package bolt

import (
	"context"

	json "github.com/goccy/go-json"
	"gitlab.com/toby3d/indieauth/internal/auth"
	"gitlab.com/toby3d/indieauth/internal/domain"
	bolt "go.etcd.io/bbolt"
)

type boltAuthRepository struct {
	db *bolt.DB
}

func NewBoltAuthRepository(db *bolt.DB) (auth.Repository, error) {
	if err := db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(domain.Login{}.Bucket())

		return err
	}); err != nil {
		return nil, err
	}

	return &boltAuthRepository{
		db: db,
	}, nil
}

func (repo *boltAuthRepository) Create(ctx context.Context, login *domain.Login) error {
	jsonLogin, err := json.Marshal(login)
	if err != nil {
		return err
	}

	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(domain.Login{}.Bucket()).Put([]byte(login.Code), jsonLogin)
	})
}

func (repo *boltAuthRepository) Get(ctx context.Context, code string) (*domain.Login, error) {
	login := new(domain.Login)

	if err := repo.db.View(func(tx *bolt.Tx) error {
		return json.Unmarshal(tx.Bucket(domain.Login{}.Bucket()).Get([]byte(code)), login)
	}); err != nil {
		return nil, err
	}

	return login, nil
}

func (repo *boltAuthRepository) Delete(ctx context.Context, code string) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(domain.Login{}.Bucket()).Delete([]byte(code))
	})
}
