package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jessevdk/go-flags"
	"golang.org/x/sync/errgroup"

	"github.com/pikocloud/pikomirror/internal/dbo"
	"github.com/pikocloud/pikomirror/internal/server"
	"github.com/pikocloud/pikomirror/internal/web"
)

//nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

type Config struct {
	Mirror struct {
		Bind    string        `long:"bind" env:"BIND" description:"Binding address" default:":8081"`
		Buffer  int           `long:"buffer" env:"BUFFER" description:"Buffer size" default:"50"`
		Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"Maximum timeout waiting for available buffer" default:"10s"`
	} `group:"Mirror server config" namespace:"mirror" env-namespace:"MIRROR"`

	Web struct {
		Bind string `long:"bind" env:"BIND" description:"Binding address" default:":8080"`
	} `group:"Web server config" namespace:"web" env-namespace:"WEB"`

	DB struct {
		Aggregation time.Duration `long:"aggregation" env:"AGGREGATION" description:"Aggregation window" default:"1s"`
		URL         string        `long:"url" env:"URL" description:"Database URL" default:"postgres://postgres:postgres@localhost/postgres"`
	} `group:"Database config" namespace:"db" env-namespace:"DB"`
}

func main() {
	var config Config

	parser := flags.NewParser(&config, flags.Default)
	parser.ShortDescription = "pikomirror"
	parser.LongDescription = fmt.Sprintf("Piko mirror - storage for mirrored requests\npikomirror %s, commit %s, built at %s by %s\nAuthor: Aleksandr Baryshnikov <owner@reddec.net>", version, commit, date, builtBy)

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if err := run(config); err != nil {
		slog.Error("run error", "error", err)
	}
}

func run(config Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := dbo.Dial(ctx, config.DB.URL)
	if err != nil {
		return fmt.Errorf("dial database: %w", err)
	}
	defer pool.Close()

	store := dbo.New(pool)

	mirror := server.NewMirror(store, config.Mirror.Buffer, config.Mirror.Timeout)

	router := chi.NewRouter()

	// TODO: add graceful stop
	router.Get("/ready", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	router.Mount("/", web.Web(store))
	router.Mount("/static/", http.FileServerFS(web.Static))

	webServer := &http.Server{
		Addr:    config.Web.Bind,
		Handler: router,
	}

	mirrorServer := &http.Server{
		Addr:    config.Mirror.Bind,
		Handler: mirror,
	}

	var wg errgroup.Group
	wg.Go(func() error {
		<-ctx.Done()
		return webServer.Close()
	})

	wg.Go(func() error {
		defer cancel()

		return mirror.Stream(ctx, config.DB.Aggregation)
	})

	wg.Go(func() error {
		defer cancel()

		return webServer.ListenAndServe()
	})

	wg.Go(func() error {
		defer cancel()

		return mirrorServer.ListenAndServe()
	})

	slog.Info("started", "version", version, "web", config.Web.Bind, "mirror", config.Mirror.Bind)
	return wg.Wait()
}
