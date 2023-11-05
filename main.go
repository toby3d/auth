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
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v9"
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
		profiles profile.UseCase
		tokens   token.UseCase
		static   fs.FS
	}

	NewAppOptions struct {
		Client   *http.Client
		Clients  client.Repository
		Sessions session.Repository
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
	config = new(domain.Config)
)

//nolint:gochecknoglobals
var (
	indieAuthClient                *domain.Client
	cpuProfilePath, memProfilePath string
)

//go:embed web/static/*
var static embed.FS

//nolint:gochecknoinits
func init() {
	flag.StringVar(&cpuProfilePath, "cpuprofile", "", "set path to saving CPU memory profile")
	flag.StringVar(&memProfilePath, "memprofile", "", "set path to saving pprof memory profile")
	flag.Parse()

	if err := env.ParseWithOptions(config, env.Options{Prefix: "AUTH_"}); err != nil {
		logger.Fatalln(err)
	}

	// NOTE(toby3d): The server instance itself can be as a client.
	rootUrl, err := url.Parse(config.Server.GetRootURL())
	if err != nil {
		logger.Fatalln(err)
	}

	cid, err := domain.ParseClientID(rootUrl.String())
	if err != nil {
		logger.Fatalln("fail to read config:", err)
	}

	indieAuthClient = &domain.Client{
		Logo:        rootUrl.JoinPath("icon.svg"),
		URL:         rootUrl,
		ID:          *cid,
		Name:        config.Name,
		RedirectURI: []*url.URL{rootUrl.JoinPath("callback")},
	}
}

//nolint:funlen,cyclop // "god object" and the entry point of all modules
func main() {
	ctx := context.Background()

	var opts NewAppOptions

	var err error
	if opts.Static, err = fs.Sub(static, filepath.Join("web", "static")); err != nil {
		logger.Fatalln(err)
	}

	switch strings.ToLower(config.Database.Type) {
	default:
		opts.Tokens = tokenmemoryrepo.NewMemoryTokenRepository()
		opts.Sessions = sessionmemoryrepo.NewMemorySessionRepository(*config)
	case "sqlite3":
		store, err := sqlx.Open("sqlite", config.Database.Path)
		if err != nil {
			logger.Fatalln(err)
		}

		if err = store.Ping(); err != nil {
			logger.Fatalf("cannot ping %s database: %v", "sqlite3", err)
		}

		opts.Tokens = tokensqlite3repo.NewSQLite3TokenRepository(store)
		opts.Sessions = sessionsqlite3repo.NewSQLite3SessionRepository(store)
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

		if config.Server.CertificateFile != "" && config.Server.KeyFile != "" {
			err = server.ListenAndServeTLS(config.Server.CertificateFile, config.Server.KeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalln("cannot listen and serve:", err)
		}
	}()

	<-done

	if err = server.Shutdown(ctx); err != nil {
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
		auth:     authucase.NewAuthUseCase(opts.Sessions, opts.Profiles, *config),
		clients:  clientucase.NewClientUseCase(opts.Clients),
		matcher:  language.NewMatcher(message.DefaultCatalog.Languages()),
		profiles: profileucase.NewProfileUseCase(opts.Profiles),
		sessions: sessionucase.NewSessionUseCase(opts.Sessions),
		tokens: tokenucase.NewTokenUseCase(tokenucase.Config{
			Config:   *config,
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
	})
	health := healthhttpdelivery.NewHandler()
	auth := authhttpdelivery.NewHandler(authhttpdelivery.NewHandlerOptions{
		Auth:     app.auth,
		Clients:  app.clients,
		Config:   *config,
		Matcher:  app.matcher,
		Profiles: app.profiles,
	})
	token := tokenhttpdelivery.NewHandler(app.tokens, *config)
	client := clienthttpdelivery.NewHandler(clienthttpdelivery.NewHandlerOptions{
		Client:  *indieAuthClient,
		Config:  *config,
		Matcher: app.matcher,
		Tokens:  app.tokens,
	})
	user := userhttpdelivery.NewHandler(app.tokens, *config)
	staticHandler := http.FileServer(http.FS(app.static))

	return http.HandlerFunc(middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		head, tail := urlutil.ShiftPath(r.URL.Path)

		switch head {
		default: // NOTE(toby3d): static or 404
			staticHandler.ServeHTTP(w, r)
		case "", "callback": // NOTE(toby3d): self-client
			client.ServeHTTP(w, r)
		case "token", "introspect", "revocation":
			token.ServeHTTP(w, r)
		case ".well-known": // NOTE(toby3d): public server config
			r.URL.Path = tail

			if head, _ = urlutil.ShiftPath(r.URL.Path); head == "oauth-authorization-server" {
				metadata.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		case "authorize":
			r.URL.Path = tail

			auth.ServeHTTP(w, r)
		case "health":
			r.URL.Path = tail

			health.ServeHTTP(w, r)
		case "userinfo":
			r.URL.Path = tail

			user.ServeHTTP(w, r)
		}
	}).Intercept(middleware.LogFmt()))
}
