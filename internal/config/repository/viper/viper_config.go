package viper

import (
	"path/filepath"

	"github.com/spf13/viper"
	"gitlab.com/toby3d/indieauth/internal/config"
)

type viperConfigRepository struct {
	viper *viper.Viper
}

func NewViperConfigRepository(v *viper.Viper) (config.Repository, error) {
	v.AddConfigPath(filepath.Join(".", "configs"))
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	for key, value := range map[string]interface{}{
		"database.client":              "bolt",
		"database.connection.filename": "data/development.db",
		"server.port":                  3000,
		"url":                          "http://127.0.0.1:3000/",
	} {
		v.SetDefault(key, value)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return &viperConfigRepository{
		viper: v,
	}, nil
}

func (v *viperConfigRepository) GetString(key string) string {
	return v.viper.GetString(key)
}
