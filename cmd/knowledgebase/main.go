package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	PAGES_DIR  string
	ASSETS_DIR string
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := Config{}

	CMD := &cobra.Command{
		Use:   "knowledgebase",
		Short: "Start a knowledgebase server",
		Long:  "Start a knowledgebase server",
		Args:  cobra.NoArgs,
	}

	CMD.PersistentFlags().IntVarP(&config.PORT, "port", "p", 80, "Port to start the server on")
	CMD.PersistentFlags().StringVarP(&config.PAGES_DIR, "docs", "d", "docs", "Path to the markdown pages")
	CMD.PersistentFlags().StringVarP(&config.ASSETS_DIR, "assets", "a", "assets", "Path to the assets directory")

	CMD.RunE = func(cmd *cobra.Command, args []string) error {
		deps := Deps{
			Port:   config.PORT,
			Pages:  os.DirFS(config.PAGES_DIR),
			Assets: os.DirFS(config.ASSETS_DIR),
		}

		kb, err := knowledgebase.New(ctx, knowledgebase.Config{
			Docs:     deps.Pages,
			Assets:   deps.Assets,
			Searcher: deps.Searcher,
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
	Pages    fs.FS
	Assets   fs.FS
	Searcher search.Searcher
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
