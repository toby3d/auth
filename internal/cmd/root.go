package cmd

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/domain"
)

//nolint: gochecknoglobals
var (
	rootCmd = &cobra.Command{
		Use:   "indieauth",
		Short: "",
		Long:  "",
	}
	client = new(domain.Client)
	config = new(domain.Config)
)

//nolint: gochecknoglobals
var configPath string

//nolint: gochecknoinits
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configPath, "config", filepath.Join(".", "config.yaml"), "config file")
	viper.BindPFlag("port", startCmd.PersistentFlags().Lookup("port"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func initConfig() {
	viper.AddConfigPath(filepath.Join(".", "configs"))
	viper.SetConfigName("config")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var err error
	if err = viper.ReadInConfig(); err != nil {
		log.Fatalf("cannot load config from file %s: %v", viper.ConfigFileUsed(), err)
	}

	if err = viper.Unmarshal(config); err != nil {
		log.Fatalln("failed to read config:", err)
	}

	// NOTE(toby3d): The server instance itself can be as a client.
	rootURL := config.Server.GetRootURL()
	client.Name = []string{config.Name}

	if client.ID, err = domain.NewClientID(rootURL); err != nil {
		log.Fatalln("fail to read config:", err)
	}

	u, logo, redirect := http.AcquireURI(), http.AcquireURI(), http.AcquireURI()
	if err = u.Parse(nil, []byte(rootURL)); err != nil {
		log.Fatalln("cannot parse client URL:", err)
	}

	u.CopyTo(logo)
	u.CopyTo(redirect)
	redirect.SetPath("/callback")
	logo.SetPath(config.Server.StaticURLPrefix + "/icon.svg")

	client.URL = []*http.URI{u}
	client.Logo = []*http.URI{logo}
	client.RedirectURI = []*http.URI{redirect}
}
