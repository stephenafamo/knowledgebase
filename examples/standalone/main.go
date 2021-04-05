package main

import (
	"context"
	"net/http"

	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase"
)

func main() {
	ctx := context.Background()
	config := knowledgebase.Config{
		Store: afero.NewBasePathFs(afero.NewOsFs(), "./docs"),
	}

	kb, err := knowledgebase.New(ctx, config)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", kb.Handler())
}
