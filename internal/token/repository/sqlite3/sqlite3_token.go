package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/token"
)

type (
	Token struct {
		CreatedAt   sql.NullTime `db:"created_at"`
		AccessToken string       `db:"access_token"`
		ClientID    string       `db:"client_id"`
		Me          string       `db:"me"`
		Scope       string       `db:"scope"`
	}

	sqlite3TokenRepository struct {
		db *sqlx.DB
	}
)

const (
	QueryTable string = `CREATE TABLE IF NOT EXISTS tokens (
		access_token TEXT UNIQUE PRIMARY KEY NOT NULL,
		client_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		me TEXT NOT NULL,
		scope TEXT
	);`

	QueryGet string = `SELECT *
		FROM tokens
		WHERE access_token=$1;`

	QueryCreate string = `INSERT INTO tokens (created_at, access_token, client_id, me, scope)
		VALUES (:created_at, :access_token, :client_id, :me, :scope);`
)

func NewSQLite3TokenRepository(db *sqlx.DB) token.Repository {
	db.MustExec(QueryTable)

	return &sqlite3TokenRepository{
		db: db,
	}
}

func (repo *sqlite3TokenRepository) Create(ctx context.Context, accessToken *domain.Token) error {
	if _, err := repo.db.NamedExecContext(ctx, QueryCreate, NewToken(accessToken)); err != nil {
		return fmt.Errorf("cannot create token record in db: %w", err)
	}

	return nil
}

func (repo *sqlite3TokenRepository) Get(ctx context.Context, accessToken string) (*domain.Token, error) {
	tkn := new(Token)
	if err := repo.db.GetContext(ctx, tkn, QueryGet, accessToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, token.ErrNotExist
		}

		return nil, fmt.Errorf("cannot find token in db: %w", err)
	}

	result := new(domain.Token)
	tkn.Populate(result)

	return result, nil
}

func NewToken(src *domain.Token) *Token {
	return &Token{
		CreatedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		AccessToken: src.AccessToken,
		ClientID:    src.ClientID.String(),
		Me:          src.Me.String(),
		Scope:       src.Scope.String(),
	}
}

func (t *Token) Populate(dst *domain.Token) {
	dst.AccessToken = t.AccessToken
	dst.ClientID, _ = domain.ParseClientID(t.ClientID)
	dst.Me, _ = domain.ParseMe(t.Me)
	dst.Scope = make(domain.Scopes, 0)

	for _, scope := range strings.Fields(t.Scope) {
		s, err := domain.ParseScope(scope)
		if err != nil {
			continue
		}

		dst.Scope = append(dst.Scope, s)
	}
}
