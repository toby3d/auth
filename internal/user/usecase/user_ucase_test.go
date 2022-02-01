package usecase_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"source.toby3d.me/website/indieauth/internal/domain"
	repository "source.toby3d.me/website/indieauth/internal/user/repository/memory"
	ucase "source.toby3d.me/website/indieauth/internal/user/usecase"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	me := domain.TestMe(t, "https://user.example.net")
	user := domain.TestUser(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, me.String()), user)

	result, err := ucase.NewUserUseCase(repository.NewMemoryUserRepository(store)).
		Fetch(context.TODO(), me)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}
