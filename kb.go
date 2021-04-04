package knowledgebase

import (
	"html/template"

	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase/search"
)

var DefaultMountPath = "/"
var DefaultDocsDir = "pages"
var DefaultAssetsDir = "assets"

type KB struct {
	Store afero.Fs // Store containing the docs and assets

	// mount path for links in the menu. Default "/"
	// Useful if the handler is to be mounted in a subdirectory of the server
	MountPath string

	// RootURL is the main application URL. Useful if the knowledgebase is part of a larger application
	// Default is the MountPath
	RootURL string

	// RootLabel is the label for the "Home" link at the top of the sidebar. Default: Home
	RootLabel string

	// MountLabel is the label for the documentation root.
	// It will not be displayed in the sidebar if empty OR if the
	// RootURL is not set or the RootURL is the same as the MountPath.
	// In these scenarios, the RootURL is the MountPath and the RootLabel will suffice
	MountLabel string

	// Directory in the store where the markdown files are
	// Default "pages"
	PagesDir string

	// Directory in the store where the referenced assets in the docs are
	// Default "assets"
	AssetsDir string

	// BaseMenu is a list of menu items that will be displayed before the
	// menu generated from the pages.
	// Example:
	// BaseMenu: []*knowledgebase.MenuItem{
	//     {
	//         Label: "Back to main site",
	//         Path:  "/",
	//     },
	//     {
	//         Label: "Login",
	//         Path:  "http://example.com/login",
	//     },
	//     {
	//         Label: "Signup",
	//         Path:  "http://example.com/signup",
	//     },
	// },
	BaseMenu []*MenuItem

	Searcher search.Searcher

	InHead, BeforeBody, AfterBody template.HTML

	templates *template.Template
	menu      []*MenuItem
}
