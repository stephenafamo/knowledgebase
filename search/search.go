package search

import (
	"context"
	"io/fs"
)

const DefaultLimit = 24

type SearchParams struct {
	Term string
}

type Searcher interface {
	// SearchDocs is used to search the index. It is passed a query and the max number
	// of results to return.
	// If limit is 0, the default limit is used.
	// A slice of post ids is returned
	SearchDocs(ctx context.Context, params SearchParams) (paths []string, err error)

	// IndexDocs indexes the entire docs directory
	// The argument of the function is the directory to index
	// This should clear all previous indexe
	IndexDocs(ctx context.Context, store fs.FS, dir string) error
}
