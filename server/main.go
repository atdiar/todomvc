package main

import (
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	const staticPath = "assets"
	const indexPath = "index.html"

	fileServer := http.FileServer(http.Dir(staticPath))

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		path = filepath.Join(staticPath, path)

		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			// file does not exist, serve index.html
			http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileServer.ServeHTTP(w, r)
	}))
} 
