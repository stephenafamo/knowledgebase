package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/afero"
)

type sqlite struct {
	db *sql.DB
}

func (s sqlite) SearchDocs(ctx context.Context, params SearchParams) ([]string, error) {
	if params.Term == "" {
		return nil, fmt.Errorf("must provide a term to search for")
	}

	args := []interface{}{}

	query := "SELECT path FROM docs WHERE docs MATCH :term"
	args = append(args, sql.Named("term", params.Term))

	query += " ORDER BY rank;"

	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("could not find docs from index: %w", err)
	}

	paths := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return paths, fmt.Errorf("error scanning id: %w", err)
		}
		paths = append(paths, id)
	}

	if err := rows.Err(); err != nil {
		return paths, fmt.Errorf("rows scan error: %w", err)
	}

	return paths, nil
}

func (s sqlite) IndexDocs(ctx context.Context, store afero.Fs, dir string) error {
	return nil
}

func NewSqlite(db *sql.DB) (sqlite, error) {
	s := sqlite{}

	columns := []string{}
	columnsWeight := []string{}

	var colData = map[string]string{
		"path UNINDEXED": "1.0", // Do not index this column
		"title":          "10.0",
		"content":        "1.0",
	}

	for name, weight := range colData {
		columns = append(columns, name)
		columnsWeight = append(columnsWeight, weight)
	}

	tx, err := db.Begin()
	if err != nil {
		return s, err
	}

	_, err = tx.Exec(`
CREATE VIRTUAL TABLE docs USING FTS5(` + strings.Join(columns, ", ") + `, tokenize = "unicode61 remove_diacritics 2");
INSERT INTO docs(docs, rank) VALUES('rank', 'bm25(` + strings.Join(columnsWeight, ", ") + `)');
`)
	if err != nil {
		return s, err
	}

	err = tx.Commit()
	if err != nil {
		return s, err
	}

	s.db = db

	return s, nil
}
