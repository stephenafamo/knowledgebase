package knowledgebase

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	images "github.com/mdigger/goldmark-images"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldHtml "github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/main.html
var mainTemplate string

//go:embed assets/build/js/app.js
var js string

//go:embed assets/build/css/app.css
var css string

func getTemplates(config Config) (*template.Template, error) {
	functions := map[string]interface{}{}

	functions["GetStyles"] = func() template.CSS {
		return template.CSS(css)
	}
	functions["GetScripts"] = func() template.JS {
		return template.JS(js)
	}
	functions["MarkdownToHTML"] = markdownToHTML
	functions["InHead"] = func() template.HTML { return config.InHead }
	functions["BeforeBody"] = func() template.HTML { return config.BeforeBody }
	functions["AfterBody"] = func() template.HTML { return config.AfterBody }

	t := template.New("Views").Funcs(functions)

	return t.Parse(string(mainTemplate))
}

func markdownToHTML(src string) (template.HTML, error) {
	imageURL := func(src string) string {
		return src
	}

	str, err := getGoldMark(src, imageURL)
	return template.HTML(str), err
}

func getGoldMark(src string, imageURL func(string) string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldHtml.WithUnsafe(),
		),
		images.NewReplacer(imageURL),
	).Convert([]byte(src), &buf); err != nil {
		return "", fmt.Errorf("could not convert MD to html: %w", err)
	}
	return buf.String(), nil
}
