//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=./web
//go:generate go install golang.org/x/text/cmd/gotext@latest
//go:generate gotext -srclang=en update -out=catalog_gen.go -lang=en,ru
package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	http "github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "modernc.org/sqlite"

	authhttpdelivery "source.toby3d.me/website/indieauth/internal/auth/delivery/http"
	authucase "source.toby3d.me/website/indieauth/internal/auth/usecase"
	clienthttpdelivery "source.toby3d.me/website/indieauth/internal/client/delivery/http"
	clientrepo "source.toby3d.me/website/indieauth/internal/client/repository/http"
	clientucase "source.toby3d.me/website/indieauth/internal/client/usecase"
	"source.toby3d.me/website/indieauth/internal/domain"
	healthhttpdelivery "source.toby3d.me/website/indieauth/internal/health/delivery/http"
	metadatahttpdelivery "source.toby3d.me/website/indieauth/internal/metadata/delivery/http"
	"source.toby3d.me/website/indieauth/internal/session"
	sessionmemoryrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	sessionsqlite3repo "source.toby3d.me/website/indieauth/internal/session/repository/sqlite3"
	"source.toby3d.me/website/indieauth/internal/ticket"
	tickethttpdelivery "source.toby3d.me/website/indieauth/internal/ticket/delivery/http"
	ticketmemoryrepo "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
	ticketsqlite3repo "source.toby3d.me/website/indieauth/internal/ticket/repository/sqlite3"
	ticketucase "source.toby3d.me/website/indieauth/internal/ticket/usecase"
	"source.toby3d.me/website/indieauth/internal/token"
	tokenhttpdelivery "source.toby3d.me/website/indieauth/internal/token/delivery/http"
	tokenmemoryrepo "source.toby3d.me/website/indieauth/internal/token/repository/memory"
	tokensqlite3repo "source.toby3d.me/website/indieauth/internal/token/repository/sqlite3"
	tokenucase "source.toby3d.me/website/indieauth/internal/token/usecase"
)

const (
	DefaultCacheDuration time.Duration = 8760 * time.Hour // NOTE(toby3d): year
	DefaultPort          int           = 3000
)

//nolint: gochecknoglobals
var (
	// NOTE(toby3d): write logs in stdout, see: https://12factor.net/logs
	logger = log.New(os.Stdout, "IndieAuth\t", log.Lmsgprefix|log.LstdFlags|log.LUTC)
	client = new(domain.Client)
	config = new(domain.Config)

	configPath     string
	cpuProfilePath string
	memProfilePath string
	enablePprof    bool
)

//nolint: gochecknoinits
func init() {
	flag.StringVar(&configPath, "config", filepath.Join(".", "config.yml"), "load specific config")
	flag.BoolVar(&enablePprof, "pprof", false, "enable pprof mode")
	flag.Parse()

	viper.AddConfigPath(filepath.Join(".", "configs"))
	viper.SetConfigName("config")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	viper.SetEnvPrefix("INDIEAUTH_")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.WatchConfig()

	var err error
	if err = viper.ReadInConfig(); err != nil {
		logger.Fatalf("cannot load config from file %s: %v", viper.ConfigFileUsed(), err)
	}

	if err = viper.Unmarshal(&config); err != nil {
		logger.Fatalln("failed to read config:", err)
	}

	// NOTE(toby3d): The server instance itself can be as a client.
	rootURL := config.Server.GetRootURL()
	client.Name = []string{config.Name}

	if client.ID, err = domain.NewClientID(rootURL); err != nil {
		logger.Fatalln("fail to read config:", err)
	}

	url, err := domain.NewURL(rootURL)
	if err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	logo, err := domain.NewURL(rootURL + config.Server.StaticURLPrefix + "/icon.svg")
	if err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	redirectURI, err := domain.NewURL(rootURL + "/callback")
	if err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	client.URL = []*domain.URL{url}
	client.Logo = []*domain.URL{logo}
	client.RedirectURI = []*domain.URL{redirectURI}
}

//nolint: funlen
func main() {
	var (
		tokens   token.Repository
		sessions session.Repository
		tickets  ticket.Repository
	)

	switch strings.ToLower(config.Database.Type) {
	case "sqlite3":
		store, err := sqlx.Open("sqlite", config.Database.Path)
		if err != nil {
			panic(err)
		}

		if err = store.Ping(); err != nil {
			logger.Fatalf("cannot ping %s database: %v", config.Database.Type, err)
		}

		tokens = tokensqlite3repo.NewSQLite3TokenRepository(store)
		sessions = sessionsqlite3repo.NewSQLite3SessionRepository(config, store)
		tickets = ticketsqlite3repo.NewSQLite3TicketRepository(store, config)
	case "memory":
		store := new(sync.Map)
		tokens = tokenmemoryrepo.NewMemoryTokenRepository(store)
		sessions = sessionmemoryrepo.NewMemorySessionRepository(config, store)
		tickets = ticketmemoryrepo.NewMemoryTicketRepository(store, config)
	default:
		log.Fatalln("unsupported database type, use 'memory' or 'sqlite3'")
	}

	go sessions.GC()

	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	httpClient := &http.Client{
		Name:               fmt.Sprintf("%s/0.1 (+%s)", config.Name, config.Server.GetAddress()),
		MaxConnDuration:    10 * time.Second,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		MaxConnWaitTimeout: 10 * time.Second,
	}
	ticketService := ticketucase.NewTicketUseCase(tickets, httpClient)
	tokenService := tokenucase.NewTokenUseCase(tokens, sessions, config)

	r := router.New()
	tickethttpdelivery.NewRequestHandler(ticketService, matcher, config).Register(r)
	healthhttpdelivery.NewRequestHandler().Register(r)
	metadatahttpdelivery.NewRequestHandler(config).Register(r)
	tokenhttpdelivery.NewRequestHandler(tokenService).Register(r)
	clienthttpdelivery.NewRequestHandler(clienthttpdelivery.NewRequestHandlerOptions{
		Client:  client,
		Config:  config,
		Matcher: matcher,
		Tokens:  tokenService,
	}).Register(r)
	authhttpdelivery.NewRequestHandler(authhttpdelivery.NewRequestHandlerOptions{
		Clients: clientucase.NewClientUseCase(clientrepo.NewHTTPClientRepository(httpClient)),
		Auth:    authucase.NewAuthUseCase(sessions, config),
		Matcher: matcher,
		Config:  config,
	}).Register(r)
	r.ServeFilesCustom(path.Join(config.Server.StaticURLPrefix, "{filepath:*}"), &http.FS{
		Root:               config.Server.StaticRootPath,
		CacheDuration:      DefaultCacheDuration,
		AcceptByteRange:    true,
		Compress:           true,
		CompressBrotli:     true,
		GenerateIndexPages: true,
	})

	if enablePprof {
		r.GET("/debug/pprof/{filepath:*}", pprofhandler.PprofHandler)
	}

	server := &http.Server{
		Name:                  fmt.Sprintf("IndieAuth/0.1 (+%s)", config.Server.GetAddress()),
		Handler:               r.Handler,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		DisableKeepalive:      true,
		ReduceMemoryUsage:     true,
		SecureErrorLogMessage: true,
		CloseOnShutdown:       true,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	if cpuProfilePath != "" {
		cpuProfile, err := os.Create(cpuProfilePath)
		if err != nil {
			logger.Fatalln("could not create CPU profile:", err)
		}
		defer cpuProfile.Close()

		if err = pprof.StartCPUProfile(cpuProfile); err != nil {
			logger.Fatalln("could not start CPU profile:", err)
		}
		defer pprof.StopCPUProfile()
	}

	go func() {
		logger.Printf("started at %s, available at %s", config.Server.GetAddress(),
			config.Server.GetRootURL())

		err := server.ListenAndServe(config.Server.GetAddress())
		if err != nil && !errors.Is(err, http.ErrConnectionClosed) {
			logger.Fatalln("cannot listen and serve:", err)
		}
	}()

	<-done

	if err := server.Shutdown(); err != nil {
		logger.Fatalln("failed shutdown of server:", err)
	}

	if memProfilePath == "" {
		return
	}

	memProfile, err := os.Create(memProfilePath)
	if err != nil {
		logger.Fatalln("could not create memory profile:", err)
	}
	defer memProfile.Close()

	runtime.GC() // NOTE(toby3d): get up-to-date statistics

	if err = pprof.WriteHeapProfile(memProfile); err != nil {
		logger.Fatalln("could not write memory profile:", err)
	}
}
