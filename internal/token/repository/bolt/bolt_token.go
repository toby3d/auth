package bolt

import (
	"context"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/token"
)

type (
	Token struct {
		AccessToken string `json:"accessToken"`
		ClientID    string `json:"clientId"`
		Me          string `json:"me"`
		Scope       string `json:"scope"`
		Type        string `json:"type"`
	}

	boltTokenRepository struct {
		db *bolt.DB
	}
)

var ErrNotExist error = xerrors.New("key not exist")

func NewBoltTokenRepository(db *bolt.DB) token.Repository {
	return &boltTokenRepository{
		db: db,
	}
}

func (repo *boltTokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	result := domain.NewToken()

	if err := repo.db.View(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		if src := tx.Bucket(Token{}.Bucket()).Get([]byte(accessToken)); src != nil {
			return new(Token).Bind(src, result)
		}

		return ErrNotExist
	}); err != nil {
		if !xerrors.Is(err, ErrNotExist) {
			return nil, errors.Wrap(err, "failed to retrieve token from storage")
		}

		return nil, nil
	}

	return result, nil
}

func (repo *boltTokenRepository) Create(ctx context.Context, accessToken *domain.Token) error {
	t, err := repo.Get(ctx, accessToken.AccessToken)
	if err != nil {
		return errors.Wrap(err, "failed to verify the existence of the token")
	}

	if t != nil {
		return token.ErrExist
	}

	return repo.Update(ctx, accessToken)
}

func (repo *boltTokenRepository) Update(ctx context.Context, accessToken *domain.Token) error {
	dto := new(Token)
	dto.Populate(accessToken)

	src, err := json.Marshal(dto)
	if err != nil {
		return errors.Wrap(err, "failed to marshal token")
	}

	if err = repo.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(dto.Bucket()).Put([]byte(dto.AccessToken), src); err != nil {
			return errors.Wrap(err, "failed to overwrite the token in the bucket")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to update the token in the repository")
	}

	return nil
}

func (repo *boltTokenRepository) Remove(ctx context.Context, accessToken string) error {
	if err := repo.db.Update(func(tx *bolt.Tx) error {
		//nolint: exhaustivestruct
		if err := tx.Bucket(Token{}.Bucket()).Delete([]byte(accessToken)); err != nil {
			return errors.Wrap(err, "failed to remove token in bucket")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to remove token from storage")
	}

	return nil
}

func (Token) Bucket() []byte { return []byte("tokens") }

func (t *Token) Populate(src *domain.Token) {
	t.AccessToken = src.AccessToken
	t.ClientID = src.ClientID
	t.Me = src.Me
	t.Scope = strings.Join(src.Scopes, " ")
	t.Type = src.Type
}

func (t *Token) Bind(src []byte, dst *domain.Token) error {
	if err := json.Unmarshal(src, t); err != nil {
		return errors.Wrap(err, "cannot unmarshal token source")
	}

	dst.AccessToken = t.AccessToken
	dst.Scopes = strings.Fields(t.Scope)
	dst.Type = t.Type
	dst.ClientID = t.ClientID
	dst.Me = t.Me

	return nil
}
