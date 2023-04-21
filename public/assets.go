//go:build !embedassets
// +build !embedassets

package public

import (
	"net/http"
)

var AssetsEmbedded = false

func Assets() (http.FileSystem, error) {
	return http.Dir("./public/build"), nil
}

func NewIndexHandler() (http.HandlerFunc, error) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("This route is not available in dev mode."))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}, nil
}
