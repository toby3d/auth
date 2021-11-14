package config

import "time"

type UseCase interface {
	GetDatabasePath() string                              // data/indieauth.db
	GetDatabaseType() string                              // bolt
	GetIndieAuthAccessTokenExpirationTime() time.Duration // time.Hour
	GetIndieAuthCodeLength() int                          // 32
	GetIndieAuthEnabled() bool                            // true
	GetIndieAuthJWTSecret() []byte                        // hackme
	GetIndieAuthJWTSigningAlgorithm() string              // RS256
	GetIndieAuthJWTSigningPrivateKeyFile() string         // jwt/private.pem
	GetIndieAuthJWTNonceLength() int                      // 22
	GetName() string                                      // IndieAuth
	GetRunMode() string                                   // dev
	GetServerAddress() string                             // 0.0.0.0:3000
	GetServerCertificate() string                         // https/cert.pem
	GetServerDomain() string                              // localhost
	GetServerEnablePPROF() bool                           // false
	GetServerHost() string                                // 0.0.0.0
	GetServerKey() string                                 // https/key.pem
	GetServerPort() int                                   // 3000
	GetServerProtocol() string                            // http
	GetServerRootURL() string                             // http://localhost:3000/
	GetServerStaticRootPath() string                      // /
	GetServerStaticURLPrefix() string                     // /static
}
