package usecase_test

import (
	"context"
	"errors"
	"path"
	"reflect"
	"sync"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/client"
	repository "source.toby3d.me/toby3d/auth/internal/client/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/client/usecase"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestDiscovery(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	testClient, localhostClient := domain.TestClient(t), domain.TestClient(t)
	localhostClient.ID, _ = domain.ParseClientID("http://localhost.toby3d.me/")

	for _, client := range []*domain.Client{testClient, localhostClient} {
		store.Store(path.Join(repository.DefaultPathPrefix, client.ID.String()), client)
	}

	for _, tc := range []struct {
		name     string
		in       *domain.Client
		out      *domain.Client
		expError error
	}{{
		name: "default",
		in:   testClient,
		out:  testClient,
	}, {
		name:     "localhost",
		in:       localhostClient,
		expError: client.ErrNotExist,
	}} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := usecase.NewClientUseCase(repository.NewMemoryClientRepository(store)).
				Discovery(context.Background(), tc.in.ID)
			if tc.expError != nil && !errors.Is(err, tc.expError) {
				t.Errorf("Discovery(%s) = %+v, want %+v", tc.in.ID, err, tc.expError)

				return
			}

			if tc.expError == nil && err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(result, tc.out) {
				t.Errorf("Discovery(%s) = %+v, want %+v", tc.in.ID, result, tc.out)
			}
		})
	}
}
