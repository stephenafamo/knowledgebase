package knowledgebase

//go:generate go install github.com/markbates/pkger/cmd/pkger
//go:generate pkger

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/markbates/pkger"
	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase/internal"
	"github.com/stephenafamo/knowledgebase/search"
)

type Menu = internal.Menu

var DefaultDocsDir = "pages"
var DefaultAssetsDir = "assets"

type KB struct {
	Store afero.Fs // Store containing the docs and assets

	PagesDir  string
	AssetsDir string
	Searcher  search.Searcher

	templates *template.Template
	menu      *Menu
}

func (ws *KB) setTemplates() error {
	functions := map[string]interface{}{}

	functions["MarkdownToHTML"] = internal.MarkdownToHTML
	functions["MenuHTML"] = internal.MenuHTML
	functions["GetStyles"] = internal.GetStyles
	functions["GetScripts"] = internal.GetScripts

	t := template.New("Views").Funcs(functions)

	path := "/templates/main.html"

	file, err := pkger.Open(path)
	if err != nil {
		err = fmt.Errorf("could not open file %q: %w", path, err)
		return err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("coud not get file contents")
		return err
	}

	ws.templates, err = t.Parse(string(contents))
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	return nil
}

func (ws *KB) Handler(ctx context.Context) (http.Handler, error) {
	var err error

	if ws.PagesDir == "" {
		ws.PagesDir = DefaultDocsDir
	}

	if ws.AssetsDir == "" {
		ws.AssetsDir = DefaultAssetsDir
	}

	if ws.Searcher != nil {
		err = ws.Searcher.IndexDocs(ctx, ws.Store, ws.PagesDir)
		if err != nil {
			return nil, fmt.Errorf("could not index docs: %w", err)
		}
	}

	err = ws.setTemplates()
	if err != nil {
		return nil, fmt.Errorf("could not set templates: %w", err)
	}

	err = ws.buildMenu()
	if err != nil {
		return nil, fmt.Errorf("could not build menu: %w", err)
	}

	httpFs := afero.NewHttpFs(ws.Store)
	fileserver := http.FileServer(httpFs.Dir(ws.AssetsDir))

	r := chi.NewRouter()
	r.Mount(
		fmt.Sprintf("/%s/", ws.AssetsDir),
		http.StripPrefix(fmt.Sprintf("/%s", ws.AssetsDir), fileserver))

	r.Mount("/", http.HandlerFunc(ws.serveDocs))

	var handler http.Handler = r

	return handler, nil
}

func (ws KB) serveDocs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "" || path == "/" {
		path = "index.md"
	}

	fullPath := filepath.Join(ws.PagesDir, path)

	file, err := ws.Store.Open(fullPath)
	if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		err = fmt.Errorf("could not open file %q: %w", path, err)
		panic(err)
	}
	if errors.Is(err, afero.ErrFileNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	defer file.Close()

	markdown, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("coud not get file contents")
		panic(err)
	}

	data := map[string]interface{}{
		"menu":    ws.menu,
		"url":     r.URL,
		"content": string(markdown),
	}

	err = ws.templates.Execute(w, data)
	if err != nil {
		panic(fmt.Errorf("could not render data: %w", err))
	}
}

func (ws *KB) buildMenu() error {
	menu := &Menu{
		Children: make([]*Menu, 0),
	}

	// Walking through embed directory
	err := afero.Walk(ws.Store, ws.PagesDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// skip hidden directories and files
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}

			// Ignore non-markdown files
			if !info.IsDir() && !strings.HasSuffix(info.Name(), ".md") {
				return nil
			}

			path = strings.TrimPrefix(path, ws.PagesDir)
			path = strings.TrimPrefix(path, "/")

			// do not add these to the menu
			if path == "" || path == "/" || path == "index.md" {
				return nil
			}

			pathParts := strings.Split(info.Name(), " ")

			if len(pathParts) < 2 {
				return fmt.Errorf("must add the order before filename for %q", path)
			}

			order, err := strconv.ParseUint(pathParts[0], 10, 64)
			if err != nil {
				return fmt.Errorf("order must be a positive integer")
			}

			name := strings.TrimSuffix(
				strings.TrimSpace(
					strings.Join(pathParts[1:], " "),
				), filepath.Ext(info.Name()),
			)

			parentMenu := menu
			splitPath := strings.Split(path, "/")
			for key, dir := range splitPath {
				if key == len(splitPath)-1 {
					continue
				}
				dirOrder, err := strconv.Atoi(strings.Split(dir, " ")[0])
				if err != nil {
					return fmt.Errorf("could not get dir order: %w", err)
				}
				parentMenu = parentMenu.Children[dirOrder]
			}

			if len(parentMenu.Children) <= int(order) {
				x := make([]*Menu, order+1)
				copy(x, parentMenu.Children)
				parentMenu.Children = x
			}
			parentMenu.Children[order] = &Menu{
				Label:    name,
				Path:     path,
				Children: make([]*Menu, 0),
			}

			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("Error walking through docs directory: %w", err)
		return err
	}

	ws.menu = menu
	return nil
}
