package bolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type (
	Token struct {
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		DeletedAt   time.Time `json:"deletedAt,omitempty"`
		Scopes      []string  `json:"scopes"`
		AccessToken string    `json:"accessToken"`
		ClientID    string    `json:"clientId"`
		Me          string    `json:"me"`
	}

	boltTokenRepository struct {
		db *bolt.DB
	}
)

func NewBoltTokenRepository(db *bolt.DB) token.Repository {
	return &boltTokenRepository{
		db: db,
	}
}

func (repo *boltTokenRepository) Create(ctx context.Context, accessToken *domain.Token) (err error) {
	t, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil && !xerrors.Is(err, token.ErrNotExist) {
		return errors.Wrap(err, "cannot get token in database")
	}

	if t != nil {
		return token.ErrExist
	}

	if err = repo.db.Update(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		bkt, err := tx.CreateBucketIfNotExists(Token{}.Bucket())
		if err != nil {
			return errors.Wrap(err, "cannot create bucket")
		}

		token := new(Token)
		token.Populate(accessToken)

		src, err := json.Marshal(token)
		if err != nil {
			return errors.Wrap(err, "cannot marshal token data")
		}

		if err = bkt.Put([]byte(token.AccessToken), src); err != nil {
			return errors.Wrap(err, "cannot put token into bucket")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to put token into database")
	}

	return nil
}

func (repo *boltTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	result := new(domain.Token)

	if err := repo.db.View(func(tx *bolt.Tx) (err error) {
		t := new(Token)

		bkt := tx.Bucket(t.Bucket())
		if bkt == nil {
			return token.ErrNotExist
		}

		src := bkt.Get([]byte(accessToken))
		if src == nil {
			return token.ErrNotExist
		}

		if err = t.Bind(src, result); err != nil {
			return errors.Wrap(err, "cannot parse token")
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to view token in database")
	}

	return result, nil
}

func (Token) Bucket() []byte { return []byte("tokens") }

func (t *Token) Populate(accessToken *domain.Token) {
	t.AccessToken = accessToken.AccessToken
	t.ClientID = accessToken.ClientID
	t.CreatedAt = time.Now().UTC().Round(time.Second)
	t.Me = accessToken.Me
	t.Scopes = make([]string, len(accessToken.Scopes))
	t.UpdatedAt = t.CreatedAt

	for i := range accessToken.Scopes {
		t.Scopes[i] = accessToken.Scopes[i]
	}
}

func (t *Token) Bind(src []byte, accessToken *domain.Token) error {
	if err := json.Unmarshal(src, t); err != nil {
		return errors.Wrap(err, "cannot unmarshal token")
	}

	accessToken.AccessToken = t.AccessToken
	accessToken.ClientID = t.ClientID
	accessToken.Me = t.Me
	accessToken.Scopes = t.Scopes

	return nil
}
