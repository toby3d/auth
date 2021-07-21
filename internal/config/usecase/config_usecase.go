package usecase

import (
	"net"

	"gitlab.com/toby3d/indieauth/internal/config"
)

type configUseCase struct {
	repo config.Repository
}

func NewConfigUseCase(repo config.Repository) config.UseCase {
	return &configUseCase{
		repo: repo,
	}
}

func (useCase *configUseCase) GetURL() string {
	return useCase.repo.GetString("url")
}

func (useCase *configUseCase) GetHost() string {
	return useCase.repo.GetString("server.host")
}

func (useCase *configUseCase) GetPort() string {
	return useCase.repo.GetString("server.port")
}

func (useCase *configUseCase) GetAddr() string {
	return net.JoinHostPort(useCase.GetHost(), useCase.GetPort())
}

func (useCase *configUseCase) GetDatabaseFileName() string {
	return useCase.repo.GetString("database.connection.filename")
}
