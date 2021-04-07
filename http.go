package knowledgebase

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func serveDocs(config Config, menu []*MenuItem, exec *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "" || path == "/" {
			path = "index.md"
		}

		fullPath := filepath.Join(config.PagesDir, path)

		file, err := config.Store.Open(fullPath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			err = fmt.Errorf("could not open file %q: %w", path, err)
			panic(err)
		}
		if errors.Is(err, fs.ErrNotExist) {
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

		heading := ""
		if path != "index.md" {
			for k, v := range strings.Split(strings.Trim(path, "/"), "/") {
				if k > 0 {
					heading += " > "
				}
				title := strings.TrimSuffix(v, ".md")
				if len(strings.Split(title, " ")) > 1 {
					title = strings.SplitN(title, " ", 2)[1]
				}

				heading += title
			}
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
