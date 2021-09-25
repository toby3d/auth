package config

type Repository interface {
	GetInt(key string) int
	GetString(key string) string
}
