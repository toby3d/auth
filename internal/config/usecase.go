package config

type UseCase interface {
	GetURL() string
	GetHost() string
	GetPort() string
	GetAddr() string
	GetDatabaseFileName() string
}
