package knowledgebase

//go:generate go install github.com/markbates/pkger/cmd/pkger
//go:generate pkger

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/markbates/pkger"
	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase/search"
)

type MenuItem struct {
	Label    string
	Path     string
	Children []*MenuItem
}

var DefaultMountPath = "/"
var DefaultDocsDir = "pages"
var DefaultAssetsDir = "assets"

type KB struct {
	Store afero.Fs // Store containing the docs and assets

	// BaseMenu is a list of menu items that will be displayed before the
	// menu generatd from the pages.
	BaseMenu []*MenuItem

	// mount path for links in the menu. Default "/"
	MountPath string

	// Directory in the store where the markdown files are
	// Default "pages"
	PagesDir string

	// Directory in the store where the referenced assets in the docs are
	// Default "assets"
	AssetsDir string

	Searcher  search.Searcher
	templates *template.Template
	menu      []*MenuItem
}

func (ws *KB) setTemplates() error {
	functions := map[string]interface{}{}

	functions["MarkdownToHTML"] = MarkdownToHTML
	functions["GetStyles"] = GetStyles
	functions["GetScripts"] = GetScripts

	t := template.New("Views").Funcs(functions)

	// We must manually list the file for pkger.Open to be able to link it
	file, err := pkger.Open("/templates/main.html")
	if err != nil {
		err = fmt.Errorf("could not open file %q: %w", "/templates/main.html", err)
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

	if ws.MountPath == "" {
		ws.MountPath = DefaultMountPath
	}

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
		log.Printf("404 for file %q", fullPath)
		return
	}
	defer file.Close()

	markdown, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("coud not get file contents")
		panic(err)
	}

	data := map[string]interface{}{
		"menuHTML":  ws.MenuHTML(r.URL.Path),
		"mountPath": ws.MountPath,
		"content":   string(markdown),
	}

	err = ws.templates.Execute(w, data)
	if err != nil {
		panic(fmt.Errorf("could not render data: %w", err))
	}
}

func (ws *KB) buildMenu() error {
	menu := &MenuItem{
		Children: make([]*MenuItem, 0),
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
				x := make([]*MenuItem, order+1)
				copy(x, parentMenu.Children)
				parentMenu.Children = x
			}
			parentMenu.Children[order] = &MenuItem{
				Label:    name,
				Path:     filepath.Join(ws.MountPath, path),
				Children: make([]*MenuItem, 0),
			}

			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("Error walking through docs directory: %w", err)
		return err
	}

	ws.menu = menu.Children
	return nil
}
