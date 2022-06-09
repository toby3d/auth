package memory_test

import (
	"context"
	"path"
	"reflect"
	"sync"
	"testing"

	repository "source.toby3d.me/toby3d/auth/internal/client/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestGet(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, client.ID.String()), client)

	result, err := repository.NewMemoryClientRepository(store).
		Get(context.Background(), client.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, client) {
		t.Errorf("Get(%s) = %+v, want %+v", client.ID, result, client)
	}
}
