package usecase_test

import (
	"context"
	"path"
	"reflect"
	"sync"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
	repository "source.toby3d.me/toby3d/auth/internal/user/repository/memory"
	ucase "source.toby3d.me/toby3d/auth/internal/user/usecase"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	me := domain.TestMe(t, "https://user.example.net")
	user := domain.TestUser(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, me.String()), user)

	result, err := ucase.NewUserUseCase(repository.NewMemoryUserRepository(store)).
		Fetch(context.Background(), me)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, user) {
		t.Errorf("Fetch(%s) = %+v, want %+v", me, result, user)
	}
}
