package main

import (
	"context"
	"net/http"
	"os"

	"github.com/stephenafamo/knowledgebase"
)

func main() {
	ctx := context.Background()
	config := knowledgebase.Config{
		Store: os.DirFS("./docs"),
	}

	kb, err := knowledgebase.New(ctx, config)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", kb)
}
