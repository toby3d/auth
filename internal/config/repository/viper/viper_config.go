package viper

import (
	"github.com/spf13/viper"

	"source.toby3d.me/website/oauth/internal/config"
)

type viperConfigRepository struct {
	viper *viper.Viper
}

func NewViperConfigRepository(v *viper.Viper) config.Repository {
	return &viperConfigRepository{
		viper: v,
	}
}

func (v *viperConfigRepository) GetString(key string) string {
	return v.viper.GetString(key)
}

func (v *viperConfigRepository) GetInt(key string) int {
	return v.viper.GetInt(key)
}

func (v *viperConfigRepository) GetBool(key string) bool {
	return v.viper.GetBool(key)
}
