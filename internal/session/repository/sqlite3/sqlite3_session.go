package sqlite3

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/jmoiron/sqlx"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/session"
)

type (
	Session struct {
		CreatedAt sql.NullTime `db:"created_at"`
		Code      string       `db:"code"`
		Data      string       `db:"data"`
	}

	sqlite3SessionRepository struct {
		db *sqlx.DB
	}
)

const (
	QueryTable string = `CREATE TABLE IF NOT EXISTS sessions (
		created_at DATETIME NOT NULL,
		code TEXT UNIQUE PRIMARY KEY NOT NULL,
		data TEXT NOT NULL
	);`

	QueryGet string = `SELECT *
		FROM sessions
		WHERE code=$1;`

	QueryCreate string = `INSERT INTO sessions (created_at, code, data)
		VALUES (:created_at, :code, :data);`

	QueryDelete string = `DELETE FROM sessions
		WHERE code=$1;`
)

func NewSQLite3SessionRepository(db *sqlx.DB) session.Repository {
	db.MustExec(QueryTable)

	return &sqlite3SessionRepository{
		db: db,
	}
}

func (repo *sqlite3SessionRepository) Create(ctx context.Context, session domain.Session) error {
	src, err := NewSession(&session)
	if err != nil {
		return fmt.Errorf("cannot encode session data for store: %w", err)
	}

	if _, err := repo.db.NamedExecContext(ctx, QueryCreate, src); err != nil {
		return fmt.Errorf("cannot create session record in db: %w", err)
	}

	return nil
}

func (repo *sqlite3SessionRepository) Get(ctx context.Context, code string) (*domain.Session, error) {
	s := new(Session) //nolint:varnamelen // cannot redaclare import
	if err := repo.db.GetContext(ctx, s, QueryGet, code); err != nil {
		return nil, fmt.Errorf("cannot find session in db: %w", err)
	}

	result := new(domain.Session)
	if err := s.Populate([]byte(s.Data), result); err != nil {
		return nil, fmt.Errorf("cannot decode session data from store: %w", err)
	}

	result.Code = code

	return result, nil
}

func (repo *sqlite3SessionRepository) GetAndDelete(ctx context.Context, code string) (*domain.Session, error) {
	s := new(Session) //nolint:varnamelen // cannot redaclare import

	tx, err := repo.db.Beginx()
	if err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = tx.GetContext(ctx, s, QueryGet, code); err != nil {
		//nolint:errcheck // deffered method
		defer tx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return nil, session.ErrNotExist
		}

		return nil, fmt.Errorf("cannot find session in db: %w", err)
	}

	if _, err = tx.ExecContext(ctx, QueryDelete, code); err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("cannot remove session from db: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := new(domain.Session)
	if err = s.Populate([]byte(s.Data), result); err != nil {
		return nil, fmt.Errorf("cannot decode session data from store: %w", err)
	}

	result.Code = code

	return result, nil
}

func (repo *sqlite3SessionRepository) GC() {}

func NewSession(src *domain.Session) (*Session, error) {
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("cannot encode data to JSON: %w", err)
	}

	return &Session{
		CreatedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		Code: src.Code,
		Data: base64.StdEncoding.EncodeToString(data),
	}, nil
}

func (t *Session) Populate(src []byte, dst *domain.Session) error {
	tmp := make([]byte, base64.StdEncoding.DecodedLen(len(src)))

	n, err := base64.StdEncoding.Decode(tmp, src)
	if err != nil {
		return fmt.Errorf("cannot decode base64 data: %w", err)
	}

	if err = json.Unmarshal(tmp[:n], dst); err != nil {
		return fmt.Errorf("cannot decode JSON data: %w", err)
	}

	return nil
}
