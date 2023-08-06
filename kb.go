package knowledgebase

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// The knowledgebase handler.
// Satisfies the http.Handler interface
// NOTE: Do not create this directly such as KB{}. Alwyas use New().
type KB struct {
	config Config
	router http.Handler
}

func New(ctx context.Context, config Config) (KB, error) {
	var err error

	DefaultMountPath := "/"

	if config.MountPath == "" {
		config.MountPath = DefaultMountPath
	}

	if config.Searcher != nil {
		err = config.Searcher.IndexDocs(ctx, config.Docs)
		if err != nil {
			return KB{}, fmt.Errorf("could not index docs: %w", err)
		}
	}

	templates, err := getTemplates(config)
	if err != nil {
		return KB{}, fmt.Errorf("could not set templates: %w", err)
	}

	menu, err := buildMenu(config)
	if err != nil {
		return KB{}, fmt.Errorf("could not build menu: %w", err)
	}

	fileserver := http.FileServer(http.FS(config.Assets))

	router := chi.NewRouter()
	router.Mount("/assets", http.StripPrefix("/assets", fileserver))
	router.Mount("/", http.HandlerFunc(serveDocs(config, menu, templates)))

	kb := KB{
		config: config,
		router: middleware.StripSlashes(router),
	}

	return kb, nil
}

// Confirm if a specific page was loaded
func (kb KB) HasPage(name string) (bool, error) {
	return fileExists(kb.config.Docs, name)
}

// Confirm if a specific asset was loaded
func (kb KB) HasAsset(name string) (bool, error) {
	return fileExists(kb.config.Assets, name)
}

// Get the http.Handler which will serve the documentation site
func (kb KB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kb.router.ServeHTTP(w, r)
}

// Check if a file or directory fileExists.
func fileExists(dir fs.FS, path string) (bool, error) {
	_, err := dir.Open(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}
