# KnowledgeBase

knowledgebase is a tool to quickly generate a knowledge base server.

There are two ways to use it.

* As a CLI tool. Run the command, visit your site. [MORE](#use-as-a-cli-tool)
* As a library. Add a docs server to your Go server. [MORE](#use-as-a-library)

The directory with your documentation needs two sub-directories.

1. A directory containing your markdown files. Default **pages**.

    In this directory, there should be a file named `index.md` which will be the file rendered at the root of your server.

    Every other page or folder must begin with a number, e.g. `01 Get Started`. This is used to order them in the sidebar menu.

2. A directory containing static assets. Default **assets**.

    In your markdown pages, you can refer to any of these as `/assets/my-static-file.jpg`.

## Use as a CLI tool

The CLI tool allows you to start a web server leading to your documentation.

To install the command line tool, you should have `Go` installed.

```sh
go get github.com/stephenafamo/knowledgebase/cmd/knowledgebase

# Check the help menu
knowledgebase -h
```

The help menu: 

```
Start a knowledgebase server

Usage:
  knowledgebase [flags]

Flags:
      --assets-dir string   Folder in the docs directory where the static assets are (default "assets")
  -d, --dir string          Docs directory (default ".")
  -h, --help                help for knowledgebase
      --pages-dir string    Folder in the docs directory where the markdown pages are (default "pages")
  -p, --port int            Port to start the server on (default 80)
```

## Use as a Library

You can use this as a library, it will return a [`http.Handler`](https://golang.org/pkg/net/http/#Handler) which you can mount on any router. There are some more options when using it this way.

```go
type KB struct {
	Store afero.Fs // Store containing the docs and assets

	// mount path for links in the menu. Default "/"
	// Useful if the handler is to be mounted in a subdirectory of the server
	MountPath string

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
}

type MenuItem struct {
	Label    string
	Path     string
	Children []*MenuItem
}
```

### Examples

Using as a standalone webserver

```go
package main

import (
	"context"
	"net/http"

	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase"
)

func main() {
	ctx := context.Background()

	kb := &knowledgebase.KB{
		Store: afero.NewBasePathFs(afero.NewOsFs(), "./docs"),
	}
	handler, err := kb.Handler(ctx)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", handler)
}
```

Mount in subdirectory with Gorrila Mux

```go
package main

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase"
)

func main() {
	ctx := context.Background()

	kb := &knowledgebase.KB{
		Store:     afero.NewBasePathFs(afero.NewOsFs(), "./docs"),
		MountPath: "/docs",
	}
	docsHandler, err := kb.Handler(ctx)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", docsHandler))

	http.ListenAndServe(":8080", r)
}
```

Mount in subdirectory with Chi

```go
package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase"
)

func main() {
	ctx := context.Background()

	kb := &knowledgebase.KB{
		Store:     afero.NewBasePathFs(afero.NewOsFs(), "./docs"),
		MountPath: "/docs",
	}
	docsHandler, err := kb.Handler(ctx)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Mount("/docs", http.StripPrefix("/docs", docsHandler))

	http.ListenAndServe(":8080", r)
}
```

## Contributing

Looking forward to pull requests.
