package knowledgebase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/afero"
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

	var DefaultMountPath = "/"
	var DefaultDocsDir = "pages"
	var DefaultAssetsDir = "assets"

	if config.MountPath == "" {
		config.MountPath = DefaultMountPath
	}

	if config.PagesDir == "" {
		config.PagesDir = DefaultDocsDir
	}

	if config.AssetsDir == "" {
		config.AssetsDir = DefaultAssetsDir
	}

	if config.Searcher != nil {
		err = config.Searcher.IndexDocs(ctx, config.Store, config.PagesDir)
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

	httpFs := afero.NewHttpFs(config.Store)
	fileserver := http.FileServer(httpFs.Dir(config.AssetsDir))

	router := chi.NewRouter()
	router.Mount(
		fmt.Sprintf("/%s/", config.AssetsDir),
		http.StripPrefix(fmt.Sprintf("/%s", config.AssetsDir), fileserver),
	)
	router.Mount("/", http.HandlerFunc(serveDocs(config, menu, templates)))

	kb := KB{
		config: config,
		router: middleware.StripSlashes(router),
	}

	return kb, nil
}

// Confirm if a specific page was loaded
func (kb KB) HasPage(name string) (bool, error) {
	return afero.Exists(afero.NewBasePathFs(kb.config.Store, kb.config.PagesDir), name)
}

// Confirm if a specific asset was loaded
func (kb KB) HasAsset(name string) (bool, error) {
	return afero.Exists(afero.NewBasePathFs(kb.config.Store, kb.config.AssetsDir), name)
}

// Get the http.Handler which will serve the documentation site
func (kb KB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kb.router.ServeHTTP(w, r)
}
