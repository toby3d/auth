//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=../../web
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/fasthttp/router"
	"github.com/spf13/viper"
	http "github.com/valyala/fasthttp"
	authdelivery "gitlab.com/toby3d/indieauth/internal/auth/delivery/http"
	authrepo "gitlab.com/toby3d/indieauth/internal/auth/repository/bolt"
	authusecase "gitlab.com/toby3d/indieauth/internal/auth/usecase"
	configrepo "gitlab.com/toby3d/indieauth/internal/config/repository/viper"
	configusecase "gitlab.com/toby3d/indieauth/internal/config/usecase"
	tokendelivery "gitlab.com/toby3d/indieauth/internal/token/delivery/http"
	tokenrepo "gitlab.com/toby3d/indieauth/internal/token/repository/bolt"
	tokenusecase "gitlab.com/toby3d/indieauth/internal/token/usecase"
	bolt "go.etcd.io/bbolt"
)

var flagConfig = flag.String("config", filepath.Join(".", "config.yml"), "set specific path to config file")

func main() {
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	v := viper.New()
	dir, _ := filepath.Split(*flagConfig)

	v.AddConfigPath(dir)

	configRepo, err := configrepo.NewViperConfigRepository(v)
	if err != nil {
		errLog.Fatal(err)
	}

	config := configusecase.NewConfigUseCase(configRepo)

	db, err := bolt.Open(config.GetDatabaseFileName(), 0666, nil)
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	authRepo, err := authrepo.NewBoltAuthRepository(db)
	if err != nil {
		errLog.Fatal(err)
	}

	tokenRepo, err := tokenrepo.NewBoltTokenRepository(db)
	if err != nil {
		errLog.Fatal(err)
	}

	authUseCase := authusecase.NewAuthUseCase(authRepo)
	tokenUseCase := tokenusecase.NewTokenUseCase(authRepo, tokenRepo)
	authHandler := authdelivery.NewAuthHandler(authUseCase)
	tokenHandler := tokendelivery.NewTokenHandler(tokenUseCase)
	r := router.New()

	r.GET("/health", func(ctx *http.RequestCtx) { ctx.SetStatusCode(http.StatusOK) })
	authHandler.Register(r)
	tokenHandler.Register(r)

	server := http.Server{
		CloseOnShutdown:       true,
		Handler:               r.Handler,
		Name:                  "IndieAuth/1.0.0 (" + config.GetURL() + ")",
		Logger:                infoLog,
		LogAllErrors:          true,
		SecureErrorLogMessage: true,
	}

	infoLog.Printf("IndieAuth started on %s", config.GetAddr())
	if err = server.ListenAndServe(config.GetAddr()); err != nil {
		errLog.Fatal(err)
	}
}
