package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui/build
var assets embed.FS

func Assets() (http.FileSystem, error) {
	fsys, err := fs.Sub(assets, "ui/build")

	if err != nil {
		return nil, err
	}

	return http.FS(fsys), nil
}
