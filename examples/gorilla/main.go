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
