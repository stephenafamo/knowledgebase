package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/stephenafamo/janus"
	"github.com/stephenafamo/knowledgebase"
	"github.com/stephenafamo/knowledgebase/search"
	"github.com/stephenafamo/orchestra"
)

type Config struct {
	PORT       int
	ROOT_DIR   string
	PAGES_DIR  string
	ASSETS_DIR string
}

func main() {
	var ctx = context.Background()
	var config = Config{}

	CMD := &cobra.Command{
		Use:   "knowledgebase",
		Short: "Start a knowledgebase server",
		Long:  "Start a knowledgebase server",
		Args:  cobra.NoArgs,
	}

	CMD.PersistentFlags().IntVarP(&config.PORT, "port", "p", 80,
		"Port to start the server on")
	CMD.PersistentFlags().StringVarP(&config.ROOT_DIR, "dir", "d", ".",
		"Docs directory")
	CMD.PersistentFlags().StringVar(&config.PAGES_DIR, "pages-dir", "pages",
		"Folder in the docs directory where the markdown pages are")
	CMD.PersistentFlags().StringVar(&config.ASSETS_DIR, "assets-dir", "assets",
		"Folder in the docs directory where the static assets are")

	CMD.RunE = func(cmd *cobra.Command, args []string) error {

		// Get the CMD dependencies
		deps, err := getDeps(config, nil)
		if err != nil {
			return fmt.Errorf("could not get deps: %w", err)
		}

		kb, err := knowledgebase.New(ctx, knowledgebase.Config{
			Store:     deps.Store,
			PagesDir:  config.PAGES_DIR,
			AssetsDir: config.ASSETS_DIR,
			Searcher:  deps.Searcher,
		})
		if err != nil {
			panic(err)
		}
		server := docsServer(deps, kb)

		err = orchestra.PlayUntilSignal(server, os.Interrupt, syscall.SIGTERM)
		if err != nil {
			panic(err)
		}
		return nil
	}

	err := CMD.Execute()
	if err != nil {
		err = fmt.Errorf("error while running command: %w", err)
		panic(err)
	}
}

type Deps struct {
	Port     int
	Store    fs.FS
	Searcher search.Searcher
}

func getDeps(config Config, db *sql.DB) (Deps, error) {
	var err error

	var deps = Deps{
		Port: config.PORT,
	}

	deps.Store, err = getStore(config)
	if err != nil {
		return deps, fmt.Errorf("could not get store: %w", err)
	}

	return deps, nil
}

func getStore(config Config) (fs.FS, error) {
	store := os.DirFS(config.ROOT_DIR)
	return store, nil
}

func docsServer(deps Deps, handler http.Handler) orchestra.Player {
	b := &janus.Broker{}

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

	log.Printf("starting docs server on port %d", deps.Port)
	return orchestra.ServerPlayer{Server: server}
}
