package knowledgebase

import (
	"html/template"
	"io/fs"

	"github.com/stephenafamo/knowledgebase/search"
)

type Config struct {
	Docs   fs.FS // Store containing the docs
	Assets fs.FS // Store containing the assets

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

	// Used to style some elements in the documentation, such as links
	PrimaryColor string

	// An optional logo that will be displayed on the sidebar
	// A link to the image is needed. Should be square. Will be used as the src
	// A data URI can be used
	Logo template.URL

	InHead, BeforeBody, AfterBody template.HTML

	// This content will be added at the end of every doc page
	// BEFORE the markdown is converted to HTML
	// A good use for this is to add markdown link references that are used in
	// multiple places. see https://spec.commonmark.org/0.29/#link-reference-definition
	SharedMarkdown string
}
