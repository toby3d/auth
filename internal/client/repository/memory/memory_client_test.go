package memory_test

import (
	"context"
	"path"
	"reflect"
	"sync"
	"testing"

	repository "source.toby3d.me/website/indieauth/internal/client/repository/memory"
	"source.toby3d.me/website/indieauth/internal/domain"
)

func TestGet(t *testing.T) {
	t.Parallel()

	client := domain.TestClient(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, client.ID.String()), client)

	result, err := repository.NewMemoryClientRepository(store).
		Get(context.TODO(), client.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, client) {
		t.Errorf("Get(%s) = %+v, want %+v", client.ID, result, client)
	}
}
