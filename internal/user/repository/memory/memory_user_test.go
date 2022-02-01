package memory_test

import (
	"context"
	"path"
	"reflect"
	"sync"
	"testing"

	"source.toby3d.me/website/indieauth/internal/domain"
	repository "source.toby3d.me/website/indieauth/internal/user/repository/memory"
)

func TestGet(t *testing.T) {
	t.Parallel()

	user := domain.TestUser(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, user.Me.String()), user)

	result, err := repository.NewMemoryUserRepository(store).Get(context.TODO(), user.Me)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, user) {
		t.Errorf("Get(%s) = %+v, want %+v", user.Me, result, user)
	}
}
