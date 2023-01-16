package usecase_test

import (
	"context"
	"reflect"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
	repository "source.toby3d.me/toby3d/auth/internal/user/repository/memory"
	ucase "source.toby3d.me/toby3d/auth/internal/user/usecase"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	user := domain.TestUser(t)
	user.Me = domain.TestMe(t, "https://user.example.net")
	users := repository.NewMemoryUserRepository()

	if err := users.Create(context.Background(), *user); err != nil {
		t.Fatal(err)
	}

	result, err := ucase.NewUserUseCase(users).Fetch(context.Background(), *user.Me)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, user) {
		t.Errorf("Fetch(%s) = %+v, want %+v", user.Me, result, user)
	}
}
