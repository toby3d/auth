package bolt

import (
	"context"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/xerrors"
	"source.toby3d.me/website/oauth/internal/model"
	"source.toby3d.me/website/oauth/internal/token"
)

type (
	Token struct {
		AccessToken string `json:"access_token"`
		ClientID    string `json:"client_id"`
		Me          string `json:"me"`
		Scope       string `json:"scope"`
		Type        string `json:"type"`
	}

	boltTokenRepository struct {
		db *bolt.DB
	}
)

var ErrNotExist error = xerrors.New("key not exist")

func NewBoltTokenRepository(db *bolt.DB) (token.Repository, error) {
	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(Token{}.Bucket())

		return err
	}); err != nil {
		return nil, err
	}

	return &boltTokenRepository{
		db: db,
	}, nil
}

func (repo *boltTokenRepository) Get(ctx context.Context, accessToken string) (*model.Token, error) {
	result := new(model.Token)
	err := repo.db.View(func(tx *bolt.Tx) error {
		if src := tx.Bucket(Token{}.Bucket()).Get([]byte(accessToken)); src != nil {
			return NewToken().Bind(src, result)
		}

		return ErrNotExist
	})
	if err != nil && !xerrors.Is(err, ErrNotExist) {
		return nil, err
	}

	if xerrors.Is(err, ErrNotExist) {
		return nil, nil
	}

	return result, nil
}

func (repo *boltTokenRepository) Create(ctx context.Context, accessToken *model.Token) error {
	t, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil {
		return errors.Wrap(err, "failed to verify the existence of the token")
	}

	if t != nil {
		return token.ErrExist
	}

	return repo.Update(ctx, accessToken)
}

func (repo *boltTokenRepository) Update(ctx context.Context, accessToken *model.Token) error {
	dto := NewToken()
	dto.Populate(accessToken)

	src, err := json.Marshal(dto)
	if err != nil {
		return errors.Wrap(err, "failed to marshal token")
	}

	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(dto.Bucket()).Put([]byte(dto.AccessToken), src)
	})
}

func (repo *boltTokenRepository) Remove(ctx context.Context, accessToken string) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(Token{}.Bucket()).Delete([]byte(accessToken))
	})
}

func NewToken() *Token {
	return new(Token)
}

func (Token) Bucket() []byte { return []byte("tokens") }

func (t *Token) Populate(src *model.Token) {
	t.AccessToken = src.AccessToken
	t.ClientID = src.ClientID
	t.Me = src.Me
	t.Scope = strings.Join(src.Scopes, " ")
	t.Type = src.Type
}

func (t *Token) Bind(src []byte, dst *model.Token) error {
	if err := json.Unmarshal(src, t); err != nil {
		return err
	}

	dst.AccessToken = t.AccessToken
	dst.ClientID = t.ClientID
	dst.Me = t.Me
	dst.Scopes = strings.Fields(t.Scope)
	dst.Type = t.Type

	return nil
}
