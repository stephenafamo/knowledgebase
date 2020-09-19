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
