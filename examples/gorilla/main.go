package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/stephenafamo/knowledgebase"
)

func main() {
	ctx := context.Background()
	config := knowledgebase.Config{
		Store:     os.DirFS("./docs"),
		MountPath: "/docs",
	}

	kb, err := knowledgebase.New(ctx, config)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", kb))

	http.ListenAndServe(":8080", r)
}
