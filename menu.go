package knowledgebase

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

type MenuItem struct {
	Label    string
	Path     string
	Children []*MenuItem
}

func buildMenu(config Config) ([]*MenuItem, error) {
	menu := &MenuItem{
		Children: make([]*MenuItem, 0),
	}

	// Walking through embed directory
	err := afero.Walk(config.Store, config.PagesDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// skip hidden directories and files
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}

			// Ignore non-markdown files
			if !info.IsDir() && !strings.HasSuffix(path, ".md") {
				return nil
			}

			path = filepath.Clean(filepath.ToSlash(path))
			path = strings.TrimPrefix(path, config.PagesDir)
			path = strings.TrimPrefix(path, "/")

			// do not add these to the menu
			if path == "" || path == "/" || filepath.Base(path) == "index.md" {
				return nil
			}

			pathParts := strings.Split(filepath.Base(path), " ")

			if len(pathParts) < 2 {
				return fmt.Errorf("must add the order before filename for %q", path)
			}

			order, err := strconv.ParseUint(pathParts[0], 10, 64)
			if err != nil {
				return fmt.Errorf("order must be a positive integer")
			}

			name := strings.TrimSuffix(
				strings.TrimSpace(
					strings.Join(pathParts[1:], " "),
				), filepath.Ext(path),
			)

			parentMenu := menu
			splitPath := strings.Split(path, "/")
			for key, dir := range splitPath {
				if key == len(splitPath)-1 {
					continue
				}
				dirOrder, err := strconv.Atoi(strings.Split(dir, " ")[0])
				if err != nil {
					return fmt.Errorf("could not get dir order: %w", err)
				}
				parentMenu = parentMenu.Children[dirOrder]
			}

			if len(parentMenu.Children) <= int(order) {
				x := make([]*MenuItem, order+1)
				copy(x, parentMenu.Children)
				parentMenu.Children = x
			}
			parentMenu.Children[order] = &MenuItem{
				Label:    name,
				Path:     filepath.ToSlash(filepath.Join(config.MountPath, path)),
				Children: make([]*MenuItem, 0),
			}

			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("Error walking through docs directory: %w", err)
		return nil, err
	}

	return menu.Children, nil
}

func menuHTML(config Config, menu []*MenuItem, currPath string) template.HTML {
	baseMenu, _ := printMenu(config, config.BaseMenu, currPath)
	mainMenu, _ := printMenu(config, menu, currPath)

	return template.HTML(baseMenu + mainMenu)
}

func printMenu(config Config, children []*MenuItem, currPath string) (markup string, isActive bool) {
	const menuClassesDefault = `flex items-center px-2 py-2 text-sm font-medium text-gray-600 group leading-5 rounded-md hover:text-gray-900 hover:bg-gray-50 focus:outline-none focus:text-gray-900 focus:bg-gray-50 transition ease-in-out duration-150`
	const menuClassesActive = `flex items-center px-2 py-2 text-sm font-medium text-gray-900 bg-gray-100 group leading-5 rounded-md hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:bg-gray-200 transition ease-in-out duration-150`

	var str string

	if children == nil {
		return str, false
	}

	var anyChildActive bool
	for _, child := range children {
		if child == nil || child.Label == "" {
			continue
		}

		selfIsActive := false
		classes := menuClassesDefault
		if filepath.Join(config.MountPath, currPath) == child.Path {
			classes = menuClassesActive
			selfIsActive = true
			anyChildActive = true
		}

		if len(child.Children) == 0 {
			str += fmt.Sprintf("<a class=%q href=%q>%s</a>", classes,
				child.Path, child.Label)
			continue
		}

		childrenMarkup, aChildIsActive := printMenu(config, child.Children, currPath)

		// It is active if either itself or any child is active
		isActive = selfIsActive || aChildIsActive

		childMarkup := fmt.Sprintf("<details class=%q> <summary>%s</summary> %s </details>\n",
			classes, child.Label, childrenMarkup)

		if isActive {
			childMarkup = fmt.Sprintf("<details open class=%q> <summary>%s</summary> %s </details>\n",
				classes, child.Label, childrenMarkup)
		}

		str += childMarkup
	}

	return str, anyChildActive
}
