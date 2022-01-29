package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/session"
)

type (
	Session struct {
		CreatedAt           sql.NullTime `db:"created_at"`
		ClientID            string       `db:"client_id"`
		Me                  string       `db:"me"`
		RedirectURI         string       `db:"redirect_uri"`
		CodeChallengeMethod string       `db:"code_challenge_method"`
		Scope               string       `db:"scope"`
		Code                string       `db:"code"`
		CodeChallenge       string       `db:"code_challenge"`
	}

	sqlite3SessionRepository struct {
		config *domain.Config
		db     *sqlx.DB
	}
)

const (
	QueryTable string = `CREATE TABLE IF NOT EXISTS sessions (
		created_at DATETIME NOT NULL,
		client_id TEXT NOT NULL,
		me TEXT NOT NULL,
		redirect_uri TEXT NOT NULL,
		code_challenge_method TEXT,
		scope TEXT,
		code TEXT UNIQUE PRIMARY KEY NOT NULL,
		code_challenge TEXT
	);`

	QueryGet string = `SELECT *
		FROM sessions
		WHERE code=$1;`

	QueryCreate string = `INSERT INTO sessions (created_at, client_id, me, redirect_uri, code_challenge_method,
			scope, code, code_challenge)
		VALUES (:created_at, :client_id, :me, :redirect_uri, :code_challenge_method, :scope, :code,
			:code_challenge);`

	QueryDelete string = `DELETE FROM sessions
		WHERE code=$1;`
)

func NewSQLite3SessionRepository(config *domain.Config, db *sqlx.DB) session.Repository {
	return &sqlite3SessionRepository{
		db: db,
	}
}

func (repo *sqlite3SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if _, err := repo.db.NamedExecContext(ctx, QueryTable+QueryCreate, NewSession(session)); err != nil {
		return fmt.Errorf("cannot create session record in db: %w", err)
	}

	return nil
}

func (repo *sqlite3SessionRepository) Get(ctx context.Context, code string) (*domain.Session, error) {
	s := new(Session)
	if err := repo.db.GetContext(ctx, s, QueryTable+QueryGet, code); err != nil {
		return nil, fmt.Errorf("cannot find session in db: %w", err)
	}

	result := new(domain.Session)
	s.Populate(result)

	return result, nil
}

func (repo *sqlite3SessionRepository) GetAndDelete(ctx context.Context, code string) (*domain.Session, error) {
	s := new(Session)

	tx, err := repo.db.Beginx()
	if err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = tx.GetContext(ctx, s, QueryTable+QueryGet, code); err != nil {
		defer tx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return nil, session.ErrNotExist
		}

		return nil, fmt.Errorf("cannot find session in db: %w", err)
	}

	if _, err = tx.ExecContext(ctx, QueryDelete, code); err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("cannot remove session from db: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := new(domain.Session)
	s.Populate(result)

	return result, nil
}

func (repo *sqlite3SessionRepository) GC() {}

func NewSession(src *domain.Session) *Session {
	return &Session{
		CreatedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		ClientID:            src.ClientID.String(),
		Code:                src.Code,
		CodeChallenge:       src.CodeChallenge,
		CodeChallengeMethod: src.CodeChallengeMethod.String(),
		Me:                  src.Me.String(),
		RedirectURI:         src.RedirectURI.String(),
		Scope:               src.Scope.String(),
	}
}

func (t *Session) Populate(dst *domain.Session) {
	dst.ClientID, _ = domain.ParseClientID(t.ClientID)
	dst.Code = t.Code
	dst.CodeChallenge = t.CodeChallenge
	dst.CodeChallengeMethod, _ = domain.ParseCodeChallengeMethod(t.CodeChallengeMethod)
	dst.Me, _ = domain.ParseMe(t.Me)
	dst.RedirectURI, _ = domain.ParseURL(t.RedirectURI)

	for _, scope := range strings.Fields(t.Scope) {
		s, err := domain.ParseScope(scope)
		if err != nil {
			continue
		}

		dst.Scope = append(dst.Scope, s)
	}
}
