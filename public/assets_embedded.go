//go:build embedassets
// +build embedassets

package public

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"text/template"
)

var AssetsEmbedded = true

//go:embed build
var assets embed.FS

func Assets() (http.FileSystem, error) {
	fsys, err := fs.Sub(assets, "build")

	if err != nil {
		return nil, err
	}

	return http.FS(fsys), nil
}

// NewIndexHandler parses and executes the webpack-built index.html
// Then returns a handler that serves that templated file
func NewIndexHandler(basePath string) (http.HandlerFunc, error) {
	// TODO remove unce ui routes are moved to root
	basePath, err := url.JoinPath(basePath, "/ui/")
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join("build", "index.html")
	p, err := assets.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(indexPath).Parse(string(p))
	if err != nil {
		return nil, fmt.Errorf("could not parse '%s' template: %q", indexPath, err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, map[string]string{
		"BaseURL": basePath,
	}); err != nil {
		return nil, fmt.Errorf("could not execute '%s' template: %q", indexPath, err)
	}
	bufBytes := buf.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		_, err := w.Write(bufBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}, nil
}
