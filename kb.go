package knowledgebase

import (
	"context"
	"fmt"
	"html/template"

	"github.com/spf13/afero"
	"github.com/stephenafamo/knowledgebase/search"
)

type knowledgebase struct {
	config    Config
	templates *template.Template
	menu      []*MenuItem
}

type Config struct {
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

	// Used to style some elements in the documentation, such as links
	PrimaryColor string

	InHead, BeforeBody, AfterBody template.HTML

	// This content will be added at the end of every doc page
	// BEFORE the markdown is converted to HTML
	// A good use for this is to add markdown link references that are used in
	// multiple places. see https://spec.commonmark.org/0.29/#link-reference-definition
	SharedMarkdown string
}

func New(ctx context.Context, config Config) (*knowledgebase, error) {
	var err error

	var DefaultMountPath = "/"
	var DefaultDocsDir = "pages"
	var DefaultAssetsDir = "assets"

	kb := &knowledgebase{config: config}

	if kb.config.MountPath == "" {
		kb.config.MountPath = DefaultMountPath
	}

	if kb.config.PagesDir == "" {
		kb.config.PagesDir = DefaultDocsDir
	}

	if kb.config.AssetsDir == "" {
		kb.config.AssetsDir = DefaultAssetsDir
	}

	if kb.config.Searcher != nil {
		err = kb.config.Searcher.IndexDocs(ctx, kb.config.Store, kb.config.PagesDir)
		if err != nil {
			return nil, fmt.Errorf("could not index docs: %w", err)
		}
	}

	err = kb.setTemplates()
	if err != nil {
		return nil, fmt.Errorf("could not set templates: %w", err)
	}

	err = kb.buildMenu()
	if err != nil {
		return nil, fmt.Errorf("could not build menu: %w", err)
	}

	return kb, nil
}
