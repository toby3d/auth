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

	"source.toby3d.me/website/indieauth/internal/auth"
	authhttpdelivery "source.toby3d.me/website/indieauth/internal/auth/delivery/http"
	authucase "source.toby3d.me/website/indieauth/internal/auth/usecase"
	"source.toby3d.me/website/indieauth/internal/client"
	clienthttpdelivery "source.toby3d.me/website/indieauth/internal/client/delivery/http"
	clienthttprepo "source.toby3d.me/website/indieauth/internal/client/repository/http"
	clientucase "source.toby3d.me/website/indieauth/internal/client/usecase"
	"source.toby3d.me/website/indieauth/internal/domain"
	healthhttpdelivery "source.toby3d.me/website/indieauth/internal/health/delivery/http"
	metadatahttpdelivery "source.toby3d.me/website/indieauth/internal/metadata/delivery/http"
	"source.toby3d.me/website/indieauth/internal/profile"
	profilehttprepo "source.toby3d.me/website/indieauth/internal/profile/repository/http"
	"source.toby3d.me/website/indieauth/internal/session"
	sessionmemoryrepo "source.toby3d.me/website/indieauth/internal/session/repository/memory"
	sessionsqlite3repo "source.toby3d.me/website/indieauth/internal/session/repository/sqlite3"
	sessionucase "source.toby3d.me/website/indieauth/internal/session/usecase"
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
	userhttpdelivery "source.toby3d.me/website/indieauth/internal/user/delivery/http"
)

type (
	App struct {
		auth     auth.UseCase
		clients  client.UseCase
		matcher  language.Matcher
		sessions session.UseCase
		tickets  ticket.UseCase
		tokens   token.UseCase
	}

	NewAppOptions struct {
		Client   *http.Client
		Clients  client.Repository
		Sessions session.Repository
		Tickets  ticket.Repository
		Tokens   token.Repository
		Profiles profile.Repository
	}
)

const (
	DefaultCacheDuration time.Duration = 8760 * time.Hour // NOTE(toby3d): year
	DefaultReadTimeout   time.Duration = 5 * time.Second
	DefaultWriteTimeout  time.Duration = 10 * time.Second
)

//nolint: gochecknoglobals
var (
	// NOTE(toby3d): write logs in stdout, see: https://12factor.net/logs
	logger          = log.New(os.Stdout, "IndieAuth\t", log.Lmsgprefix|log.LstdFlags|log.LUTC)
	config          = new(domain.Config)
	indieAuthClient = new(domain.Client)

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

	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(".", "configs"))
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

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
	indieAuthClient.Name = []string{config.Name}

	if indieAuthClient.ID, err = domain.ParseClientID(rootURL); err != nil {
		logger.Fatalln("fail to read config:", err)
	}

	url, err := domain.ParseURL(rootURL)
	if err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	logo, err := domain.ParseURL(rootURL + config.Server.StaticURLPrefix + "/icon.svg")
	if err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	redirectURI, err := domain.ParseURL(rootURL + "/callback")
	if err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	indieAuthClient.URL = []*domain.URL{url}
	indieAuthClient.Logo = []*domain.URL{logo}
	indieAuthClient.RedirectURI = []*domain.URL{redirectURI}
}

//nolint: funlen, cyclop // "god object" and the entry point of all modules
func main() {
	var opts NewAppOptions

	switch strings.ToLower(config.Database.Type) {
	case "sqlite3":
		store, err := sqlx.Open("sqlite", config.Database.Path)
		if err != nil {
			panic(err)
		}

		if err = store.Ping(); err != nil {
			logger.Fatalf("cannot ping %s database: %v", "sqlite3", err)
		}

		opts.Tokens = tokensqlite3repo.NewSQLite3TokenRepository(store)
		opts.Sessions = sessionsqlite3repo.NewSQLite3SessionRepository(store)
		opts.Tickets = ticketsqlite3repo.NewSQLite3TicketRepository(store, config)
	case "memory":
		store := new(sync.Map)
		opts.Tokens = tokenmemoryrepo.NewMemoryTokenRepository(store)
		opts.Sessions = sessionmemoryrepo.NewMemorySessionRepository(store, config)
		opts.Tickets = ticketmemoryrepo.NewMemoryTicketRepository(store, config)
	default:
		log.Fatalln("unsupported database type, use 'memory' or 'sqlite3'")
	}

	go opts.Sessions.GC()

	//nolint: exhaustivestruct // too many options
	opts.Client = &http.Client{
		Name:         fmt.Sprintf("%s/0.1 (+%s)", config.Name, config.Server.GetAddress()),
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
	}
	opts.Clients = clienthttprepo.NewHTTPClientRepository(opts.Client)
	opts.Profiles = profilehttprepo.NewHTPPClientRepository(opts.Client)

	r := router.New() //nolint: varnamelen
	NewApp(opts).Register(r)
	//nolint: exhaustivestruct// too many options
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

	//nolint: exhaustivestruct
	server := &http.Server{
		Name:                  fmt.Sprintf("IndieAuth/0.1 (+%s)", config.Server.GetAddress()),
		Handler:               r.Handler,
		ReadTimeout:           DefaultReadTimeout,
		WriteTimeout:          DefaultWriteTimeout,
		DisableKeepalive:      true,
		ReduceMemoryUsage:     true,
		SecureErrorLogMessage: true,
		CloseOnShutdown:       true,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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

func NewApp(opts NewAppOptions) *App {
	return &App{
		auth:     authucase.NewAuthUseCase(opts.Sessions, opts.Profiles, config),
		clients:  clientucase.NewClientUseCase(opts.Clients),
		matcher:  language.NewMatcher(message.DefaultCatalog.Languages()),
		sessions: sessionucase.NewSessionUseCase(opts.Sessions),
		tickets:  ticketucase.NewTicketUseCase(opts.Tickets, opts.Client, config),
		tokens: tokenucase.NewTokenUseCase(tokenucase.Config{
			Config:   config,
			Profiles: opts.Profiles,
			Sessions: opts.Sessions,
			Tokens:   opts.Tokens,
		}),
	}
}

func (app *App) Register(r *router.Router) {
	tickethttpdelivery.NewRequestHandler(app.tickets, app.matcher, config).Register(r)
	healthhttpdelivery.NewRequestHandler().Register(r)
	metadatahttpdelivery.NewRequestHandler(&domain.Metadata{
		Issuer:                indieAuthClient.ID,
		AuthorizationEndpoint: domain.MustParseURL(indieAuthClient.ID.String() + "authorize"),
		TokenEndpoint:         domain.MustParseURL(indieAuthClient.ID.String() + "token"),
		TicketEndpoint:        domain.MustParseURL(indieAuthClient.ID.String() + "ticket"),
		MicropubEndpoint:      nil,
		MicrosubEndpoint:      nil,
		IntrospectionEndpoint: domain.MustParseURL(indieAuthClient.ID.String() + "introspect"),
		RevocationEndpoint:    domain.MustParseURL(indieAuthClient.ID.String() + "revocation"),
		UserinfoEndpoint:      domain.MustParseURL(indieAuthClient.ID.String() + "userinfo"),
		ServiceDocumentation:  domain.MustParseURL("https://indieauth.net/source/"),
		IntrospectionEndpointAuthMethodsSupported: []string{"Bearer"},
		RevocationEndpointAuthMethodsSupported:    []string{"none"},
		ScopesSupported: domain.Scopes{
			domain.ScopeBlock,
			domain.ScopeChannels,
			domain.ScopeCreate,
			domain.ScopeDelete,
			domain.ScopeDraft,
			domain.ScopeEmail,
			domain.ScopeFollow,
			domain.ScopeMedia,
			domain.ScopeMute,
			domain.ScopeProfile,
			domain.ScopeRead,
			domain.ScopeUpdate,
		},
		ResponseTypesSupported: []domain.ResponseType{
			domain.ResponseTypeCode,
			domain.ResponseTypeID,
		},
		GrantTypesSupported: []domain.GrantType{
			domain.GrantTypeAuthorizationCode,
			domain.GrantTypeTicket,
		},
		CodeChallengeMethodsSupported: []domain.CodeChallengeMethod{
			domain.CodeChallengeMethodMD5,
			domain.CodeChallengeMethodPLAIN,
			domain.CodeChallengeMethodS1,
			domain.CodeChallengeMethodS256,
			domain.CodeChallengeMethodS512,
		},
		AuthorizationResponseIssParameterSupported: true,
	}).Register(r)
	tokenhttpdelivery.NewRequestHandler(app.tokens, app.tickets, config).Register(r)
	clienthttpdelivery.NewRequestHandler(clienthttpdelivery.NewRequestHandlerOptions{
		Client:  indieAuthClient,
		Config:  config,
		Matcher: app.matcher,
		Tokens:  app.tokens,
	}).Register(r)
	authhttpdelivery.NewRequestHandler(authhttpdelivery.NewRequestHandlerOptions{
		Auth:    app.auth,
		Clients: app.clients,
		Config:  config,
		Matcher: app.matcher,
	}).Register(r)
	userhttpdelivery.NewRequestHandler(app.tokens, config).Register(r)
}
