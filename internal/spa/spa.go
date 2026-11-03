// Package spa serves a built single-page app (the Vite output in web/dist).
package spa

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// Handler serves files from dir. Any path that isn't a real file gets
// index.html, so client-side routes like /posts/3/edit work on reload.
func Handler(dir string) http.Handler {
	root := os.DirFS(dir)
	files := http.FileServerFS(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		if name != "" && name != "index.html" {
			if info, err := fs.Stat(root, name); err == nil && !info.IsDir() {
				// Vite puts content-hashed files in assets/, so they never change.
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}

		index, err := fs.ReadFile(root, "index.html")
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				http.Error(w, "frontend not built: run `npm run build` in web/ (or use `npm run dev` on :5173)", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(index)
	})
}
