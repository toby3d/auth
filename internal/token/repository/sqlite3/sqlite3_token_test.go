package sqlite3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/testing/sqltest"
	repository "source.toby3d.me/website/indieauth/internal/token/repository/sqlite3"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	token := domain.TestToken(t)
	require.NoError(t, repository.NewSQLite3TokenRepository(db).Create(context.Background(), token))

	results := make([]*repository.Token, 0)
	require.NoError(t, db.Select(&results, "SELECT * FROM tokens;"))
	require.Len(t, results, 1)

	result := new(domain.Token)
	results[0].Populate(result)

	assert.Equal(t, token.AccessToken, result.AccessToken)
}

func TestGet(t *testing.T) {
	t.Parallel()

	db, cleanup := sqltest.Open(t)
	t.Cleanup(cleanup)

	token := domain.TestToken(t)
	_, err := db.NamedExec(repository.QueryTable+repository.QueryCreate, repository.NewToken(token))
	require.NoError(t, err)

	result, err := repository.NewSQLite3TokenRepository(db).Get(context.Background(), token.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, result.AccessToken)
}
