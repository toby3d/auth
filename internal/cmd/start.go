package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/spf13/cobra"
	http "github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	clienthttpdelivery "source.toby3d.me/website/oauth/internal/client/delivery/http"
	clientrepo "source.toby3d.me/website/oauth/internal/client/repository/http"
	clientucase "source.toby3d.me/website/oauth/internal/client/usecase"
	healthhttpdelivery "source.toby3d.me/website/oauth/internal/health/delivery/http"
	metadatahttpdelivery "source.toby3d.me/website/oauth/internal/metadata/delivery/http"
	tickethttpdelivery "source.toby3d.me/website/oauth/internal/ticket/delivery/http"
	ticketucase "source.toby3d.me/website/oauth/internal/ticket/usecase"
	userrepo "source.toby3d.me/website/oauth/internal/user/repository/http"
	userucase "source.toby3d.me/website/oauth/internal/user/usecase"
)

const (
	DefaultCacheDuration time.Duration = 8760 * time.Hour // NOTE(toby3d): year
	DefaultPort          int           = 3000
)

//nolint: gochecknoglobals
var startCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Long:  "",
	Run:   startServer,
}

//nolint: gochecknoglobals
var (
	cpuProfilePath string
	memProfilePath string
	enablePprof    bool
)

//nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().IntP("port", "p", DefaultPort, "port to run server on")
	startCmd.PersistentFlags().BoolVar(&enablePprof, "pprof", false, "enable pprof mode")
	startCmd.PersistentFlags().StringVar(&cpuProfilePath, "cpuprofile", "", "write cpu profile to file")
	startCmd.PersistentFlags().StringVar(&memProfilePath, "memprofile", "", "write memory profile to file")
}

func startServer(cmd *cobra.Command, args []string) {
	store, err := bolt.Open(config.Database.Path, os.ModePerm, nil)
	if err != nil {
		log.Fatalln("failed to open database connection:", err)
	}
	defer store.Close()

	httpClient := &http.Client{
		Name: fmt.Sprintf("%s/0.1 (+%s)", config.Name, config.Server.GetAddress()),
	}
	r := router.New()
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())

	r.ServeFilesCustom(path.Join(config.Server.StaticURLPrefix, "{filepath:*}"), &http.FS{
		Root:               config.Server.StaticRootPath,
		CacheDuration:      DefaultCacheDuration,
		AcceptByteRange:    true,
		Compress:           true,
		CompressBrotli:     true,
		GenerateIndexPages: true,
	})
	healthhttpdelivery.NewRequestHandler().Register(r)
	metadatahttpdelivery.NewRequestHandler(config).Register(r)
	clienthttpdelivery.NewRequestHandler(config, client, matcher).Register(r)
	tickethttpdelivery.NewRequestHandler(
		ticketucase.NewTicketUseCase(httpClient),
		userucase.NewUserUseCase(userrepo.NewHTTPUserRepository(httpClient)),
	).Register(r)

	if enablePprof {
		r.GET("/debug/pprof/{filepath:*}", pprofhandler.PprofHandler)
	}

	server := &http.Server{
		CloseOnShutdown:  true,
		DisableKeepalive: true,
		Handler:          r.Handler,
		Logger:           log.New(os.Stdout, config.Name+"\t", log.Lmsgprefix|log.LstdFlags|log.LUTC),
		Name:             fmt.Sprintf("%s/0.1 (+%s)", config.Name, config.Server.GetAddress()),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if cpuProfilePath != "" {
		cpuProfile, err := os.Create(cpuProfilePath)
		if err != nil {
			log.Fatalln("could not create CPU profile:", err)
		}
		defer cpuProfile.Close()

		if err = pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Fatalln("could not start CPU profile:", err)
		}
		defer pprof.StopCPUProfile()
	}

	go func() {
		server.Logger.Printf(
			"started at %s, available at %s",
			config.Server.GetAddress(),
			config.Server.GetRootURL(),
		)

		err := server.ListenAndServe(config.Server.GetAddress())
		if err != nil && !errors.Is(err, http.ErrConnectionClosed) {
			log.Fatalln("cannot listen and serve:", err)
		}
	}()

	<-done

	if err = server.Shutdown(); err != nil {
		log.Fatalln("failed shutdown of server:", err)
	}

	if memProfilePath == "" {
		return
	}

	memProfile, err := os.Create(memProfilePath)
	if err != nil {
		log.Fatalln("could not create memory profile:", err)
	}
	defer memProfile.Close()

	runtime.GC() // NOTE(toby3d): get up-to-date statistics

	if err = pprof.WriteHeapProfile(memProfile); err != nil {
		log.Fatalln("could not write memory profile:", err)
	}
}
