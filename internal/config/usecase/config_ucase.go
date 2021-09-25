package usecase

import (
	"net"

	"source.toby3d.me/website/oauth/internal/config"
)

type configUseCase struct {
	repo config.Repository
}

func NewConfigUseCase(repo config.Repository) config.UseCase {
	return &configUseCase{
		repo: repo,
	}
}

func (useCase *configUseCase) URL() string {
	return useCase.repo.GetString("url")
}

func (useCase *configUseCase) Host() string {
	return useCase.repo.GetString("server.host")
}

func (useCase *configUseCase) Port() int {
	return useCase.repo.GetInt("server.port")
}

func (useCase *configUseCase) Addr() string {
	return net.JoinHostPort(useCase.repo.GetString("server.host"), useCase.repo.GetString("server.port"))
}

func (useCase *configUseCase) DBFileName() string {
	return useCase.repo.GetString("database.connection.filename")
}
