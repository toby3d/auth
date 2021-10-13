package bolt

import (
	"context"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type boltTokenRepository struct {
	db *bolt.DB
}

var ErrNotExist error = errors.New("token not exist")

var DefaultBucket = []byte("tokens") //nolint: gochecknoglobals

func NewBoltTokenRepository(db *bolt.DB) token.Repository {
	return &boltTokenRepository{
		db: db,
	}
}

func (repo *boltTokenRepository) Create(ctx context.Context, accessToken *domain.Token) error {
	find, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil {
		return errors.Wrap(err, "cannot check token in database")
	}

	if find != nil {
		return token.ErrExist
	}

	if err = repo.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(DefaultBucket)
		if err != nil {
			return errors.Wrap(err, "cannot create bucket")
		}

		err = bkt.Put([]byte(accessToken.AccessToken), []byte(accessToken.Expiry.Format(time.RFC3339)))
		if err != nil {
			return errors.Wrap(err, "cannot put token into bucket")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to batch token in database")
	}

	return nil
}

func (repo *boltTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	result := &domain.Token{
		Expiry:      time.Time{},
		Scopes:      []string{},
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ClientID:    "",
		Me:          "",
	}

	if err := repo.db.View(func(tx *bolt.Tx) (err error) {
		bkt := tx.Bucket(DefaultBucket)
		if bkt == nil {
			return ErrNotExist
		}

		expiry := bkt.Get([]byte(accessToken))
		if expiry == nil {
			return ErrNotExist
		}

		if result.Expiry, err = time.Parse(time.RFC3339, string(expiry)); err != nil {
			return errors.Wrap(err, "cannot parse expiry date")
		}

		return nil
	}); err != nil {
		if xerrors.Is(err, ErrNotExist) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to view token in database")
	}

	return result, nil
}
