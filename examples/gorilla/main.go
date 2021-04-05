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
	config := knowledgebase.Config{
		Store:     afero.NewBasePathFs(afero.NewOsFs(), "./docs"),
		MountPath: "/docs",
	}

	kb, err := knowledgebase.New(ctx, config)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", kb.Handler()))

	http.ListenAndServe(":8080", r)
}
