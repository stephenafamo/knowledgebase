package knowledgebase

//go:generate go install github.com/markbates/pkger/cmd/pkger
//go:generate pkger

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// mount path for links in the menu. Default "/"
	// Useful if the handler is to be mounted in a subdirectory of the server
	MountPath string

	// RootURL is the main application URL. Useful if the knowledgebase is part of a larger application
	// Default is the MountPath
	RootURL string

	// RootLabel is the label for the "Home" link at the top of the sidebar. Default: Home
	RootLabel string

	// MountLabel is the label for the documentation root.
	// It will not be displayed in the sidebar if empty OR if the
	// RootURL is not set or the RootURL is the same as the MountPath.
	// In these scenarios, the RootURL is the MountPath and the RootLabel will suffice
	MountLabel string

	// Directory in the store where the markdown files are
	// Default "pages"
	PagesDir string

	// Directory in the store where the referenced assets in the docs are
	// Default "assets"
	AssetsDir string

	// BaseMenu is a list of menu items that will be displayed before the
	// menu generated from the pages.
	// Example:
	// BaseMenu: []*knowledgebase.MenuItem{
	//     {
	//         Label: "Back to main site",
	//         Path:  "/",
	//     },
	//     {
	//         Label: "Login",
	//         Path:  "http://example.com/login",
	//     },
	//     {
	//         Label: "Signup",
	//         Path:  "http://example.com/signup",
	//     },
	// },
	BaseMenu []*MenuItem

	Searcher search.Searcher

	InHead, BeforeBody, AfterBody template.HTML

	templates *template.Template
	menu      []*MenuItem
}

func (ws *KB) setTemplates() error {
	functions := map[string]interface{}{}

	functions["MarkdownToHTML"] = MarkdownToHTML
	functions["GetStyles"] = GetStyles
	functions["GetScripts"] = GetScripts
	functions["InHead"] = func() template.HTML { return ws.InHead }
	functions["BeforeBody"] = func() template.HTML { return ws.BeforeBody }
	functions["AfterBody"] = func() template.HTML { return ws.AfterBody }

	t := template.New("Views").Funcs(functions)

	// We must manually list the file for pkger.Open to be able to link it
	file, err := pkger.Open("github.com/stephenafamo/knowledgebase:/templates/main.html")
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
		"menuHTML": kb.MenuHTML(r.URL.Path),
		"kb":       kb,
		"content":  string(markdown),
	}

	err = kb.templates.Execute(w, data)
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
			if !info.IsDir() && !strings.HasSuffix(path, ".md") {
				return nil
			}

			path = filepath.Clean(filepath.ToSlash(path))
			path = strings.TrimPrefix(path, ws.PagesDir)
			path = strings.TrimPrefix(path, "/")

			// do not add these to the menu
			if path == "" || path == "/" || filepath.Base(path) == "index.md" {
				return nil
			}

			pathParts := strings.Split(filepath.Base(path), " ")

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
				), filepath.Ext(path),
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
				Path:     filepath.ToSlash(filepath.Join(ws.MountPath, path)),
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
