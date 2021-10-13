package viper_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	repository "source.toby3d.me/website/oauth/internal/config/repository/viper"
)

func TestGetString(t *testing.T) {
	t.Parallel()

	v := viper.New()
	_ = v.MergeConfigMap(map[string]interface{}{
		"testing": map[string]interface{}{
			"sample": "text",
			"answer": 42,
		},
	})

	repo := repository.NewViperConfigRepository(v)
	assert.Equal(t, "text", repo.GetString("testing.sample"))
	assert.Equal(t, "42", repo.GetString("testing.answer"))
}

func TestGetInt(t *testing.T) {
	t.Parallel()

	v := viper.New()
	_ = v.MergeConfigMap(map[string]interface{}{
		"testing": map[string]interface{}{
			"answer": 42,
		},
	})

	assert.Equal(t, 42, repository.NewViperConfigRepository(v).GetInt("testing.answer"))
}

func TestGetBool(t *testing.T) {
	t.Parallel()

	v := viper.New()
	_ = v.MergeConfigMap(map[string]interface{}{
		"testing": map[string]interface{}{
			"answer":   42,
			"enabled":  true,
			"disabled": false,
		},
	})

	assert.False(t, repository.NewViperConfigRepository(v).GetBool("testing.answer"))
	assert.True(t, repository.NewViperConfigRepository(v).GetBool("testing.enabled"))
	assert.False(t, repository.NewViperConfigRepository(v).GetBool("testing.disabled"))
}
