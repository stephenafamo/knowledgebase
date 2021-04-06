package knowledgebase

import (
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

func (kb *knowledgebase) Handler() http.Handler {
	httpFs := afero.NewHttpFs(kb.config.Store)
	fileserver := http.FileServer(httpFs.Dir(kb.config.AssetsDir))

	r := chi.NewRouter()
	r.Mount(
		fmt.Sprintf("/%s/", kb.config.AssetsDir),
		http.StripPrefix(fmt.Sprintf("/%s", kb.config.AssetsDir), fileserver),
	)

	r.Mount("/", http.HandlerFunc(kb.serveDocs))

	var handler http.Handler = middleware.StripSlashes(r)

	return handler
}

func (kb knowledgebase) serveDocs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "" || path == "/" {
		path = "index.md"
	}

	fullPath := filepath.Join(kb.config.PagesDir, path)

	file, err := kb.config.Store.Open(fullPath)
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
	if path == "index.md" {
		heading = ""
	}

	if len(strings.Split(heading, " ")) > 1 {
		heading = strings.SplitN(heading, " ", 2)[1]
	}

	data := map[string]interface{}{
		"heading":      heading,
		"menuHTML":     kb.menuHTML(r.URL.Path),
		"config":       kb.config,
		"primaryColor": kb.config.PrimaryColor,
		"content":      string(markdown) + "\n\n" + kb.config.SharedMarkdown,
	}

	err = kb.templates.Execute(w, data)
	if err != nil {
		panic(fmt.Errorf("could not render data: %w", err))
	}
}
