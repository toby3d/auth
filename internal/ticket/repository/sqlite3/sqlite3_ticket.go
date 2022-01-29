package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/ticket"
	"source.toby3d.me/website/indieauth/internal/token"
)

type (
	Ticket struct {
		CreatedAt sql.NullTime `db:"created_at"`
		Resource  string       `db:"resource"`
		Subject   string       `db:"subject"`
		Ticket    string       `db:"ticket"`
	}

	sqlite3TicketRepository struct {
		config *domain.Config
		db     *sqlx.DB
	}
)

const (
	QueryTable string = `CREATE TABLE IF NOT EXISTS tickets (
		created_at DATETIME NOT NULL,
		resource TEXT NOT NULL,
		subject TEXT NOT NULL,
		ticket TEXT UNIQUE PRIMARY KEY NOT NULL
	);`

	QueryGet string = `SELECT *
		FROM tickets
		WHERE ticket=$1;`

	QueryCreate string = `INSERT INTO tickets (created_at, resource, subject, ticket)
		VALUES (:created_at, :resource, :subject, :ticket);`

	QueryDelete string = `DELETE FROM tickets
		WHERE ticket=$1;`
)

func NewSQLite3TicketRepository(db *sqlx.DB, config *domain.Config) ticket.Repository {
	return &sqlite3TicketRepository{
		config: config,
		db:     db,
	}
}

func (repo *sqlite3TicketRepository) Create(ctx context.Context, t *domain.Ticket) error {
	if _, err := repo.db.NamedExecContext(ctx, QueryTable+QueryCreate, NewTicket(t)); err != nil {
		return fmt.Errorf("cannot create token record in db: %w", err)
	}

	return nil
}

func (repo *sqlite3TicketRepository) GetAndDelete(ctx context.Context, ticket string) (*domain.Ticket, error) {
	t := new(Ticket)

	tx, err := repo.db.Beginx()
	if err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = tx.GetContext(ctx, t, QueryTable+QueryGet, ticket); err != nil {
		defer tx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return nil, token.ErrNotExist
		}

		return nil, fmt.Errorf("cannot find ticket in db: %w", err)
	}

	if _, err = tx.ExecContext(ctx, QueryDelete, ticket); err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("cannot remove ticket from db: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := new(domain.Ticket)
	t.Populate(result)

	return result, nil
}

func (repo *sqlite3TicketRepository) GC() {}

func NewTicket(src *domain.Ticket) *Ticket {
	return &Ticket{
		CreatedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		Resource: src.Resource.String(),
		Subject:  src.Subject.String(),
		Ticket:   src.Ticket,
	}
}

func (t *Ticket) Populate(dst *domain.Ticket) {
	dst.Ticket = t.Ticket
	dst.Subject, _ = domain.ParseMe(t.Subject)
	dst.Resource, _ = domain.NewURL(t.Resource)
}
