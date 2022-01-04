package memory

import (
	"context"
	"sync"

	"gitlab.com/toby3d/indieauth/internal/auth"
	"gitlab.com/toby3d/indieauth/internal/domain"
)

type memoryAuthRepository struct {
	logins *sync.Map
}

func NewMemoryAuthRepository() auth.Repository {
	return &memoryAuthRepository{
		logins: new(sync.Map),
	}
}

func (repo *memoryAuthRepository) Create(ctx context.Context, login *domain.Login) error {
	repo.logins.Store(login.Code, login)

	return nil
}

func (repo *memoryAuthRepository) Get(ctx context.Context, code string) (*domain.Login, error) {
	login, ok := repo.logins.LoadAndDelete(code)
	if !ok {
		return nil, nil
	}

	return login.(*domain.Login), nil
}

func (repo *memoryAuthRepository) Delete(ctx context.Context, code string) error {
	repo.logins.Delete(code)

	return nil
}
