package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	repository "source.toby3d.me/toby3d/auth/internal/client/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/client/usecase"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

func TestDiscovery(t *testing.T) {
	t.Parallel()

	testClient := domain.TestClient(t)
	clients := repository.NewMemoryClientRepository()

	if err := clients.Create(context.Background(), *testClient); err != nil {
		t.Fatal(err)
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
	}} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := usecase.NewClientUseCase(clients).
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
