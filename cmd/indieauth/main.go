//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=../../web
package main

import (
	"flag"
	"log"
	gohttp "net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/fasthttp/router"
	"github.com/spf13/viper"
	http "github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"

	configrepo "source.toby3d.me/website/oauth/internal/config/repository/viper"
	configucase "source.toby3d.me/website/oauth/internal/config/usecase"
	healthdelivery "source.toby3d.me/website/oauth/internal/health/delivery/http"
	tokendelivery "source.toby3d.me/website/oauth/internal/token/delivery/http"
	tokenrepo "source.toby3d.me/website/oauth/internal/token/repository/bolt"
	tokenucase "source.toby3d.me/website/oauth/internal/token/usecase"
)

//nolint: gochecknoglobals
var (
	flagConfig     = flag.String("config", filepath.Join(".", "config.yml"), "set specific path to config file")
	flagCpuProfile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	flagMemProfile = flag.String("memprofile", "", "write memory profile to `file`")
)

//nolint: funlen
func main() {
	flag.Parse()

	if *flagCpuProfile != "" || *flagMemProfile != "" {
		go log.Println(gohttp.ListenAndServe("localhost:6060", nil))
	}

	if *flagCpuProfile != "" {
		cpuProfile, err := os.Create(*flagCpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer cpuProfile.Close()

		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	v := viper.New()
	v.SetDefault("url", "/")
	v.SetDefault("database", map[string]interface{}{
		"client": "bolt",
		"connection": map[string]interface{}{
			"filename": filepath.Join(".", "development.db"),
		},
	})
	v.SetDefault("server", map[string]interface{}{
		"host": "127.0.0.1",
		"port": 3000, //nolint: gomnd
	})
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	dir, file := filepath.Split(*flagConfig)
	if file != "" {
		ext := filepath.Ext(file)
		v.SetConfigName(strings.TrimSuffix(file, ext))
		v.SetConfigType(ext[1:])
	}

	v.AddConfigPath(dir)
	v.AddConfigPath(filepath.Join(".", "configs"))
	v.AddConfigPath(".")

	r := router.New()
	cfg := configucase.NewConfigUseCase(configrepo.NewViperConfigRepository(v))

	db, err := bolt.Open(cfg.DBFileName(), os.ModePerm, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(tokenrepo.Token{}.Bucket())

		return err
	}); err != nil {
		log.Fatalln(err)
	}

	tokendelivery.NewRequestHandler(tokenucase.NewTokenUseCase(tokenrepo.NewBoltTokenRepository(db))).Register(r)
	healthdelivery.NewRequestHandler().Register(r)

	server := &http.Server{
		Handler:          r.Handler,
		Name:             "IndieAuth/1.0.0 (" + cfg.URL() + ")",
		DisableKeepalive: true,
		CloseOnShutdown:  true,
		// TODO(toby3d): Logger
	}

	if err := server.ListenAndServe(cfg.Addr()); err != nil {
		log.Fatalln(err)
	}

	if *flagMemProfile == "" {
		return
	}

	memProfile, err := os.Create(*flagMemProfile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer memProfile.Close()

	runtime.GC() // NOTE(toby3d): get up-to-date statistics
	if err := pprof.WriteHeapProfile(memProfile); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
