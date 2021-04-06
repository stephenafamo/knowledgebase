package knowledgebase

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func serveDocs(config Config, menu []*MenuItem, exec *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "" || path == "/" {
			path = "index.md"
		}

		fullPath := filepath.Join(config.PagesDir, path)

		file, err := config.Store.Open(fullPath)
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
		if path == "index.md" {
			heading = ""
		}

		if len(strings.Split(heading, " ")) > 1 {
			heading = strings.SplitN(heading, " ", 2)[1]
		}

		data := map[string]interface{}{
			"heading":      heading,
			"menuHTML":     menuHTML(config, menu, r.URL.Path),
			"config":       config,
			"primaryColor": config.PrimaryColor,
			"content":      string(markdown) + "\n\n" + config.SharedMarkdown,
		}

		err = exec.Execute(w, data)
		if err != nil {
			panic(fmt.Errorf("could not render data: %w", err))
		}
	}
}
