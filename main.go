//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=./web
//go:generate go install golang.org/x/text/cmd/gotext@master
//go:generate gotext -srclang=en update -out=catalog_gen.go -lang=en,ru
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "modernc.org/sqlite"

	"source.toby3d.me/toby3d/auth/internal/auth"
	authhttpdelivery "source.toby3d.me/toby3d/auth/internal/auth/delivery/http"
	authucase "source.toby3d.me/toby3d/auth/internal/auth/usecase"
	"source.toby3d.me/toby3d/auth/internal/client"
	clienthttpdelivery "source.toby3d.me/toby3d/auth/internal/client/delivery/http"
	clienthttprepo "source.toby3d.me/toby3d/auth/internal/client/repository/http"
	clientucase "source.toby3d.me/toby3d/auth/internal/client/usecase"
	"source.toby3d.me/toby3d/auth/internal/domain"
	healthhttpdelivery "source.toby3d.me/toby3d/auth/internal/health/delivery/http"
	metadatahttpdelivery "source.toby3d.me/toby3d/auth/internal/metadata/delivery/http"
	"source.toby3d.me/toby3d/auth/internal/middleware"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilehttprepo "source.toby3d.me/toby3d/auth/internal/profile/repository/http"
	profileucase "source.toby3d.me/toby3d/auth/internal/profile/usecase"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionmemoryrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
	sessionsqlite3repo "source.toby3d.me/toby3d/auth/internal/session/repository/sqlite3"
	sessionucase "source.toby3d.me/toby3d/auth/internal/session/usecase"
	"source.toby3d.me/toby3d/auth/internal/ticket"
	tickethttpdelivery "source.toby3d.me/toby3d/auth/internal/ticket/delivery/http"
	ticketmemoryrepo "source.toby3d.me/toby3d/auth/internal/ticket/repository/memory"
	ticketsqlite3repo "source.toby3d.me/toby3d/auth/internal/ticket/repository/sqlite3"
	ticketucase "source.toby3d.me/toby3d/auth/internal/ticket/usecase"
	"source.toby3d.me/toby3d/auth/internal/token"
	tokenhttpdelivery "source.toby3d.me/toby3d/auth/internal/token/delivery/http"
	tokenmemoryrepo "source.toby3d.me/toby3d/auth/internal/token/repository/memory"
	tokensqlite3repo "source.toby3d.me/toby3d/auth/internal/token/repository/sqlite3"
	tokenucase "source.toby3d.me/toby3d/auth/internal/token/usecase"
	"source.toby3d.me/toby3d/auth/internal/urlutil"
	userhttpdelivery "source.toby3d.me/toby3d/auth/internal/user/delivery/http"
)

type (
	App struct {
		auth     auth.UseCase
		clients  client.UseCase
		matcher  language.Matcher
		sessions session.UseCase
		tickets  ticket.UseCase
		profiles profile.UseCase
		tokens   token.UseCase
		static   fs.FS
	}

	NewAppOptions struct {
		Client   *http.Client
		Clients  client.Repository
		Sessions session.Repository
		Tickets  ticket.Repository
		Tokens   token.Repository
		Profiles profile.Repository
		Static   fs.FS
	}
)

const (
	DefaultReadTimeout  time.Duration = 5 * time.Second
	DefaultWriteTimeout time.Duration = 10 * time.Second
)

//nolint:gochecknoglobals
var (
	// NOTE(toby3d): write logs in stdout, see: https://12factor.net/logs
	logger = log.New(os.Stdout, "IndieAuth\t", log.Lmsgprefix|log.LstdFlags|log.LUTC)
	// NOTE(toby3d): read configuration from environment, see: https://12factor.net/config
	config          = new(domain.Config)
	indieAuthClient = &domain.Client{
		ID:          domain.ClientID{},
		Logo:        make([]*url.URL, 1),
		RedirectURI: make([]*url.URL, 1),
		URL:         make([]*url.URL, 1),
		Name:        make([]string, 0),
	}
)

//nolint:gochecknoglobals
var cpuProfilePath, memProfilePath string

//go:embed assets/*
var staticFS embed.FS

//nolint:gochecknoinits
func init() {
	flag.StringVar(&cpuProfilePath, "cpuprofile", "", "set path to saving CPU memory profile")
	flag.StringVar(&memProfilePath, "memprofile", "", "set path to saving pprof memory profile")
	flag.Parse()

	if err := env.Parse(config, env.Options{
		Environment:     nil,
		OnSet:           nil,
		Prefix:          "INDIEAUTH_",
		RequiredIfNoDef: false,
		TagName:         "",
	}); err != nil {
		logger.Fatalln(err)
	}

	// NOTE(toby3d): The server instance itself can be as a client.
	rootURL := config.Server.GetRootURL()
	indieAuthClient.Name = []string{config.Name}

	cid, err := domain.ParseClientID(rootURL)
	if err != nil {
		logger.Fatalln("fail to read config:", err)
	}

	indieAuthClient.ID = *cid

	if indieAuthClient.URL[0], err = url.Parse(rootURL); err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	if indieAuthClient.Logo[0], err = url.Parse(rootURL + "icon.svg"); err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}

	if indieAuthClient.RedirectURI[0], err = url.Parse(rootURL + "callback"); err != nil {
		logger.Fatalln("cannot parse root URL as client URL:", err)
	}
}

//nolint:funlen,cyclop // "god object" and the entry point of all modules
func main() {
	ctx := context.Background()

	var opts NewAppOptions

	var err error
	if opts.Static, err = fs.Sub(staticFS, "assets"); err != nil {
		logger.Fatalln(err)
	}

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
		opts.Tokens = tokenmemoryrepo.NewMemoryTokenRepository()
		opts.Sessions = sessionmemoryrepo.NewMemorySessionRepository(*config)
		opts.Tickets = ticketmemoryrepo.NewMemoryTicketRepository(*config)
	default:
		log.Fatalln("unsupported database type, use 'memory' or 'sqlite3'")
	}

	go opts.Sessions.GC()

	opts.Client = new(http.Client)
	opts.Clients = clienthttprepo.NewHTTPClientRepository(opts.Client)
	opts.Profiles = profilehttprepo.NewHTPPClientRepository(opts.Client)

	app := NewApp(opts)

	server := &http.Server{
		Addr:              config.Server.GetAddress(),
		BaseContext:       nil,
		ConnContext:       nil,
		ConnState:         nil,
		ErrorLog:          logger,
		Handler:           app.Handler(),
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		ReadHeaderTimeout: 0,
		ReadTimeout:       DefaultReadTimeout,
		TLSConfig:         nil,
		TLSNextProto:      nil,
		WriteTimeout:      DefaultWriteTimeout,
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

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalln("cannot listen and serve:", err)
		}
	}()

	<-done

	if err := server.Shutdown(ctx); err != nil {
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
		static:   opts.Static,
		auth:     authucase.NewAuthUseCase(opts.Sessions, opts.Profiles, config),
		clients:  clientucase.NewClientUseCase(opts.Clients),
		matcher:  language.NewMatcher(message.DefaultCatalog.Languages()),
		profiles: profileucase.NewProfileUseCase(opts.Profiles),
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

// TODO(toby3d): move module middlewares to here.
//
//nolint:funlen
func (app *App) Handler() http.Handler {
	//nolint:exhaustivestruct
	metadata := metadatahttpdelivery.NewHandler(&domain.Metadata{
		Issuer:                indieAuthClient.ID.URL(),
		AuthorizationEndpoint: indieAuthClient.ID.URL().JoinPath("authorize"),
		TokenEndpoint:         indieAuthClient.ID.URL().JoinPath("token"),
		TicketEndpoint:        indieAuthClient.ID.URL().JoinPath("ticket"),
		MicropubEndpoint:      nil,
		MicrosubEndpoint:      nil,
		IntrospectionEndpoint: indieAuthClient.ID.URL().JoinPath("introspect"),
		RevocationEndpoint:    indieAuthClient.ID.URL().JoinPath("revocation"),
		UserinfoEndpoint:      indieAuthClient.ID.URL().JoinPath("userinfo"),
		ServiceDocumentation: &url.URL{
			Scheme: "https",
			Host:   "indieauth.net",
			Path:   "/source/",
		},
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
	}).Handler()
	health := healthhttpdelivery.NewHandler().Handler()
	auth := authhttpdelivery.NewHandler(authhttpdelivery.NewHandlerOptions{
		Auth:     app.auth,
		Clients:  app.clients,
		Config:   *config,
		Matcher:  app.matcher,
		Profiles: app.profiles,
	}).Handler()
	token := tokenhttpdelivery.NewHandler(app.tokens, app.tickets, config).Handler()
	client := clienthttpdelivery.NewHandler(clienthttpdelivery.NewHandlerOptions{
		Client:  *indieAuthClient,
		Config:  *config,
		Matcher: app.matcher,
		Tokens:  app.tokens,
	}).Handler()
	user := userhttpdelivery.NewHandler(app.tokens, config).Handler()
	ticket := tickethttpdelivery.NewHandler(app.tickets, app.matcher, *config).Handler()
	static := http.FileServer(http.FS(app.static))

	return http.HandlerFunc(middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var head string
		head, r.URL.Path = urlutil.ShiftPath(r.URL.Path)

		switch head {
		case "", "callback", "token", "introspect", "revocation":
			if r.URL.Path != "/" {
				r.URL = r.URL.JoinPath(head, r.URL.Path)
			} else {
				r.URL = r.URL.JoinPath(head)
			}
		}

		switch head {
		default:
			static.ServeHTTP(w, r)
		case "", "callback":
			client.ServeHTTP(w, r)
		case "token", "introspect", "revocation":
			token.ServeHTTP(w, r)
		case ".well-known":
			if head, _ = urlutil.ShiftPath(r.URL.Path); head == "oauth-authorization-server" {
				metadata.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		case "authorize":
			auth.ServeHTTP(w, r)
		case "health":
			health.ServeHTTP(w, r)
		case "userinfo":
			user.ServeHTTP(w, r)
		case "ticket":
			ticket.ServeHTTP(w, r)
		}
	}) /*.Intercept(middleware.LogFmt())*/)
}
