package config

type Repository interface {
	GetString(key string) string
}
