package config

type UseCase interface {
	Addr() string
	DBFileName() string
	Host() string
	Port() int
	URL() string
}
