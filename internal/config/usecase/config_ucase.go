package usecase

import (
	"net"
	"path/filepath"
	"time"

	"github.com/valyala/fasttemplate"

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

func (useCase *configUseCase) GetName() string {
	return useCase.repo.GetString("name")
}

func (useCase *configUseCase) GetRunMode() string {
	return useCase.repo.GetString("runMode")
}

func (useCase *configUseCase) GetServerProtocol() string {
	return useCase.repo.GetString("server.protocol")
}

func (useCase *configUseCase) GetServerDomain() string {
	return useCase.repo.GetString("server.domain")
}

func (useCase *configUseCase) GetServerRootURL() string {
	t := fasttemplate.New(useCase.repo.GetString("server.rootUrl"), "{{", "}}")

	data := make(map[string]interface{})
	for _, key := range []string{
		"domain",
		"httpAddr",
		"httpPort",
		"protocol",
	} {
		data[key] = useCase.repo.GetString("server." + key)
	}

	return t.ExecuteString(data)
}

func (useCase *configUseCase) GetServerStaticURLPrefix() string {
	return useCase.repo.GetString("server.staticUrlPrefix")
}

func (useCase *configUseCase) GetServerHost() string {
	return useCase.repo.GetString("server.host")
}

func (useCase *configUseCase) GetServerPort() int {
	return useCase.repo.GetInt("server.port")
}

func (useCase *configUseCase) GetServerAddress() string {
	return net.JoinHostPort(useCase.repo.GetString("server.host"),
		useCase.repo.GetString("server.port"))
}

func (useCase *configUseCase) GetServerCertificate() string {
	return filepath.Clean(useCase.repo.GetString("server.certFile"))
}

func (useCase *configUseCase) GetServerKey() string {
	return filepath.Clean(useCase.repo.GetString("server.keyFile"))
}

func (useCase *configUseCase) GetServerStaticRootPath() string {
	return useCase.repo.GetString("server.staticRootPath")
}

func (useCase *configUseCase) GetServerEnablePPROF() bool {
	return useCase.repo.GetBool("server.enablePprof")
}

func (useCase *configUseCase) GetDatabaseType() string {
	return useCase.repo.GetString("database.type")
}

func (useCase *configUseCase) GetDatabasePath() string {
	return filepath.Clean(useCase.repo.GetString("database.path"))
}

func (useCase *configUseCase) GetIndieAuthEnabled() bool {
	return useCase.repo.GetBool("indieauth.enabled")
}

func (useCase *configUseCase) GetIndieAuthAccessTokenExpirationTime() time.Duration {
	return time.Duration(useCase.repo.GetInt("indieauth.accessTokenExpirationTime")) * time.Second
}

func (useCase *configUseCase) GetIndieAuthJWTSigningAlgorithm() string {
	return useCase.repo.GetString("indieauth.jwtSigningAlgorithm")
}

func (useCase *configUseCase) GetIndieAuthJWTSecret() string {
	return useCase.repo.GetString("indieauth.jwtSecret")
}

func (useCase *configUseCase) GetIndieAuthJWTSigningPrivateKeyFile() string {
	return filepath.Clean(useCase.repo.GetString("indieauth.jwtSigningPrivateKeyFile"))
}
