// Package studio serves the embedded browser client for the avatar API.
package studio

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

// Handler returns the standalone studio. Its client uses only the public HTTP API.
func Handler() http.Handler {
	files, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err) // The embedded directory is verified at compile time.
	}
	server := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/", "/app.js", "/view.mjs", "/styles.css":
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "same-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
			server.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}
