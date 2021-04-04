package knowledgebase

import (
	"fmt"
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

func (ws *KB) Has(name string) (bool, error) {
	return afero.Exists(afero.NewBasePathFs(ws.Store, ws.PagesDir), name)
}

func (ws *KB) buildMenu() error {
	menu := &MenuItem{
		Children: make([]*MenuItem, 0),
	}

	// Walking through embed directory
	err := afero.Walk(ws.Store, ws.PagesDir,
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
			path = strings.TrimPrefix(path, ws.PagesDir)
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
				Path:     filepath.ToSlash(filepath.Join(ws.MountPath, path)),
				Children: make([]*MenuItem, 0),
			}

			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("Error walking through docs directory: %w", err)
		return err
	}

	ws.menu = menu.Children
	return nil
}

func (ws KB) printMenu(children []*MenuItem, currPath string) string {
	const menuClassesDefault = `flex items-center px-2 py-2 text-sm font-medium text-gray-600 group leading-5 rounded-md hover:text-gray-900 hover:bg-gray-50 focus:outline-none focus:text-gray-900 focus:bg-gray-50 transition ease-in-out duration-150`
	const menuClassesActive = `flex items-center px-2 py-2 text-sm font-medium text-gray-900 bg-gray-100 group leading-5 rounded-md hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:bg-gray-200 transition ease-in-out duration-150`

	var str string

	if children == nil {
		return str
	}

	for _, child := range children {
		if child == nil || child.Label == "" {
			continue
		}

		classes := menuClassesDefault
		if filepath.Join(ws.MountPath, currPath) == child.Path {
			classes = menuClassesActive
		}

		if len(child.Children) == 0 {
			str += fmt.Sprintf("<a class=%q href=%q>%s</a>", classes,
				child.Path, child.Label)
			continue
		}

		str += fmt.Sprintf(
			"<details open class=%q> <summary>%s</summary> %s </details>\n",
			classes, child.Label, ws.printMenu(child.Children, currPath))
	}

	return str
}
