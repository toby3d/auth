package config

type Repository interface {
	GetBool(key string) bool
	GetInt(key string) int
	GetString(key string) string
}
