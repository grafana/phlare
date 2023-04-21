//go:build embedassets
// +build embedassets

package public

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
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

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join("build", "index.html")
	// TODO: read this at startup
	p, err := assets.ReadFile(indexPath)
	if err != nil {
		fmt.Println("err", err)
		panic("missing file")
		// TODO: Handle error as appropriate for the application.
	}
	w.Write(p)
}
