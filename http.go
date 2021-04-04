package knowledgebase

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/afero"
)

func (kb *KB) Handler(ctx context.Context) (http.Handler, error) {
	var err error

	if kb.MountPath == "" {
		kb.MountPath = DefaultMountPath
	}

	if kb.PagesDir == "" {
		kb.PagesDir = DefaultDocsDir
	}

	if kb.AssetsDir == "" {
		kb.AssetsDir = DefaultAssetsDir
	}

	if kb.Searcher != nil {
		err = kb.Searcher.IndexDocs(ctx, kb.Store, kb.PagesDir)
		if err != nil {
			return nil, fmt.Errorf("could not index docs: %w", err)
		}
	}

	err = kb.setTemplates()
	if err != nil {
		return nil, fmt.Errorf("could not set templates: %w", err)
	}

	err = kb.buildMenu()
	if err != nil {
		return nil, fmt.Errorf("could not build menu: %w", err)
	}

	httpFs := afero.NewHttpFs(kb.Store)
	fileserver := http.FileServer(httpFs.Dir(kb.AssetsDir))

	r := chi.NewRouter()
	r.Mount(
		fmt.Sprintf("/%s/", kb.AssetsDir),
		http.StripPrefix(fmt.Sprintf("/%s", kb.AssetsDir), fileserver),
	)

	r.Mount("/", http.HandlerFunc(kb.serveDocs))

	var handler http.Handler = middleware.StripSlashes(r)

	return handler, nil
}

func (kb KB) serveDocs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "" || path == "/" {
		path = "index.md"
	}

	fullPath := filepath.Join(kb.PagesDir, path)

	file, err := kb.Store.Open(fullPath)
	if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		err = fmt.Errorf("could not open file %q: %w", path, err)
		panic(err)
	}
	if errors.Is(err, afero.ErrFileNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		log.Printf("404 for file %q", fullPath)
		return
	}
	defer file.Close()

	markdown, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("coud not get file contents")
		panic(err)
	}

	filename := filepath.Base(path)
	heading := strings.TrimSuffix(filename, ".md")

	if len(strings.Split(heading, " ")) > 1 {
		heading = strings.SplitN(heading, " ", 2)[1]
	}

	data := map[string]interface{}{
		"heading":  heading,
		"menuHTML": kb.menuHTML(r.URL.Path),
		"kb":       kb,
		"content":  string(markdown),
	}

	err = kb.templates.Execute(w, data)
	if err != nil {
		panic(fmt.Errorf("could not render data: %w", err))
	}
}
