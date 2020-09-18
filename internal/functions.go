package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/markbates/pkger"
	images "github.com/mdigger/goldmark-images"
	"github.com/yuin/goldmark"
)

func MarkdownToHTML(src string) (template.HTML, error) {

	imageURL := func(src string) string {
		return src
	}

	str, err := getGoldMark(src, imageURL)
	return template.HTML(str), err
}

func getGoldMark(src string, imageURL func(string) string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.New(
		images.NewReplacer(imageURL),
	).Convert([]byte(src), &buf); err != nil {
		return "", fmt.Errorf("could not convert MD to html: %w", err)
	}
	return buf.String(), nil
}

func GetScripts() (template.JS, error) {
	js := template.JS("")

	path := "github.com/stephenafamo/knowledgebase:/assets/build/js/app.js"

	file, err := pkger.Open(path)
	if err != nil {
		err = fmt.Errorf("could not open file %q: %w", path, err)
		return js, err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("coud not get file contents")
		return js, err
	}

	js = template.JS(string(contents))
	return js, nil
}

func GetStyles() (template.CSS, error) {
	styles := template.CSS("")

	path := "github.com/stephenafamo/knowledgebase:/assets/build/css/app.css"

	file, err := pkger.Open(path)
	if err != nil {
		err = fmt.Errorf("could not open file %q: %w", path, err)
		return styles, err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("coud not get file contents")
		return styles, err
	}

	styles = template.CSS(string(contents))
	return styles, nil
}

func MenuHTML(menu *Menu, currPath string) template.HTML {
	return template.HTML(fmt.Sprintf("%s", printChlidren(menu.Children, currPath)))
}

func printChlidren(children []*Menu, currPath string) string {
	var str string

	if children == nil {
		return str
	}

	for _, child := range children {
		if child == nil || child.Label == "" {
			continue
		}

		classes := menuClasses
		if currPath == child.Path {
			classes = menuClassesA
		}

		if len(child.Children) == 0 {
			str += fmt.Sprintf("<a class=%q href=%q>%s</a>", classes, "/"+child.Path, child.Label)
			continue
		}

		str += fmt.Sprintf(
			"<details class=%q> <summary>%s</summary> %s </details>\n",
			classes, child.Label, printChlidren(child.Children, currPath))
	}
	return str
}

var menuClasses = `flex items-center px-2 py-2 text-sm font-medium text-gray-600 group leading-5 rounded-md hover:text-gray-900 hover:bg-gray-50 focus:outline-none focus:text-gray-900 focus:bg-gray-50 transition ease-in-out duration-150`
var menuClassesA = `flex items-center px-2 py-2 text-sm font-medium text-gray-900 bg-gray-100 group leading-5 rounded-md hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:bg-gray-200 transition ease-in-out duration-150`
