package memory_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"source.toby3d.me/website/indieauth/internal/domain"
	repository "source.toby3d.me/website/indieauth/internal/user/repository/memory"
)

func TestGet(t *testing.T) {
	t.Parallel()

	user := domain.TestUser(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, user.Me.String()), user)

	result, err := repository.NewMemoryUserRepository(store).Get(context.TODO(), user.Me)
	require.NoError(t, err)
	assert.Equal(t, user, result)
}
