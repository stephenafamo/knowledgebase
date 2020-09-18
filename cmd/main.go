package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/afero"
	"github.com/stephenafamo/janus"
	"github.com/stephenafamo/janus/monitor"
	jSentry "github.com/stephenafamo/janus/monitor/sentry"
	"github.com/stephenafamo/knowledgebase"
	"github.com/stephenafamo/knowledgebase/internal"
	"github.com/stephenafamo/knowledgebase/search"
	"github.com/stephenafamo/orchestra"
)

func main() {
	var ctx = context.Background()
	var config = internal.Config{}

	// Load env variables from a .env file if present
	err := godotenv.Overload("./config/.env")
	if err != nil {
		// Ignore error if file is not present
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
	}

	err = envconfig.Process(ctx, &config)
	if err != nil {
		panic(err)
	}

	// // Connect to search db
	// var searchDB *sql.DB
	// if config.TESTING {
	// if err := os.Remove("test.db"); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
	// panic(err)
	// }
	// searchDB, err = sql.Open("sqlite3", "file:test.db?_fk=1")
	// } else {
	// searchDB, err = sql.Open("sqlite3", "file::memory:?_fk=1&cache=shared&mode=memory")
	// searchDB.SetMaxOpenConns(1) // necessary for in-memory sqlite
	// }
	// if err != nil {
	// panic(err)
	// }
	// defer searchDB.Close()

	// Get the CMD dependencies
	// deps, err := getDeps(config, searchDB)
	deps, err := getDeps(config, nil)
	if err != nil {
		panic(err)
	}
	defer deps.Monitor.Flush(time.Second * 5)

	handler, err := (&knowledgebase.KB{
		Store:    deps.Store,
		Searcher: deps.Searcher,
	}).Handler(ctx)
	if err != nil {
		panic(err)
	}
	server := docsServer(deps, handler)

	err = orchestra.PlayUntilSignal(server, os.Interrupt, syscall.SIGTERM)
	if err != nil {
		panic(err)
	}
}

type Deps struct {
	Port     int
	Monitor  monitor.Monitor
	Store    afero.Fs
	Searcher search.Searcher
}

func getDeps(config internal.Config, db *sql.DB) (Deps, error) {
	var err error

	var deps = Deps{
		Port: config.PORT,
	}

	deps.Monitor, err = getMonitor(config)
	if err != nil {
		return deps, fmt.Errorf("could not get monitor: %w", err)
	}

	deps.Store, err = getStore(config)
	if err != nil {
		return deps, fmt.Errorf("could not get store: %w", err)
	}

	// deps.Searcher, err = search.NewSqlite(db)
	// if err != nil {
	// return deps, fmt.Errorf("could not get searher: %w", err)
	// }

	return deps, nil
}

func getStore(config internal.Config) (afero.Fs, error) {
	store := afero.NewBasePathFs(afero.NewOsFs(), config.ROOT_DIR)

	return store, nil
}

func getMonitor(config internal.Config) (monitor.Monitor, error) {
	options := sentry.ClientOptions{
		Dsn:   config.SENTRY_DSN,
		Debug: true,
	}

	// Add logging during testing
	if config.TESTING {
		options.Integrations = func(in []sentry.Integration) []sentry.Integration {
			return append(in, jSentry.LoggingIntegration{
				Logger:        sentryLogger{},
				SupressErrors: config.TESTING,
			})
		}
	}

	// Get the sentry client
	client, err := sentry.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("could not create sentry client: %w", err)
	}

	hub := sentry.NewHub(client, sentry.NewScope())
	return jSentry.Sentry{Hub: hub}, nil
}

type sentryLogger struct{}

func (sentryLogger) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf(format, a...)
}

func docsServer(deps Deps, handler http.Handler) orchestra.Player {
	b := &janus.Broker{
		Monitor: deps.Monitor,
	}

	mids := b.RecommendedMiddlewares()
	midLen := len(mids)
	for i := midLen - 1; i >= 0; i-- {
		handler = mids[i](handler)
	}

	server := &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(deps.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 15,
		Handler:      handler,
	}

	return orchestra.ServerPlayer{Server: server}

}
