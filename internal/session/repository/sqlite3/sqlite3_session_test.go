package sqlite3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	repository "source.toby3d.me/website/indieauth/internal/session/repository/sqlite3"
	"source.toby3d.me/website/indieauth/internal/testing/sqltest"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	session := domain.TestSession(t)
	require.NoError(t, repository.NewSQLite3SessionRepository(domain.TestConfig(t), db).
		Create(context.Background(), session))

	results := make([]*repository.Session, 0)
	require.NoError(t, db.Select(&results, "SELECT * FROM sessions"))
	require.Len(t, results, 1)

	result := new(domain.Session)
	results[0].Populate(result)

	assert.Equal(t, session.Code, result.Code)
}

func TestGet(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	session := domain.TestSession(t)
	_, err := db.NamedExec(repository.QueryTable+repository.QueryCreate, repository.NewSession(session))
	require.NoError(t, err)

	result, err := repository.NewSQLite3SessionRepository(domain.TestConfig(t), db).
		Get(context.Background(), session.Code)
	require.NoError(t, err)
	assert.Equal(t, session.Code, result.Code)
}

func TestGetAndDelete(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	session := domain.TestSession(t)
	_, err := db.NamedExec(repository.QueryTable+repository.QueryCreate, repository.NewSession(session))
	require.NoError(t, err)

	result, err := repository.NewSQLite3SessionRepository(domain.TestConfig(t), db).
		GetAndDelete(context.Background(), session.Code)
	require.NoError(t, err)
	assert.Equal(t, session.Code, result.Code)

	assert.Error(t, db.Get(result, repository.QueryGet, session.Code), "session MUST be destroyed after successful"+" query")
}
