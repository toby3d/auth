package config

import "time"

type UseCase interface {
	GetDatabasePath() string
	GetDatabaseType() string
	GetIndieAuthAccessTokenExpirationTime() time.Duration
	GetIndieAuthEnabled() bool
	GetIndieAuthJWTSecret() string
	GetIndieAuthJWTSigningAlgorithm() string
	GetIndieAuthJWTSigningPrivateKeyFile() string
	GetName() string
	GetRunMode() string
	GetServerAddress() string
	GetServerCertificate() string
	GetServerDomain() string
	GetServerEnablePPROF() bool
	GetServerHost() string
	GetServerKey() string
	GetServerPort() int
	GetServerProtocol() string
	GetServerRootURL() string
	GetServerStaticRootPath() string
	GetServerStaticURLPrefix() string
}
