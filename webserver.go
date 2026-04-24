package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func (a *App) webHandler() http.Handler {
	sub, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			a.routes().ServeHTTP(w, r)
			return
		}
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "." || p == "" {
			http.ServeFileFS(w, r, sub, "index.html")
			return
		}
		if _, err := fs.Stat(sub, p); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFileFS(w, r, sub, "index.html")
	})
}
