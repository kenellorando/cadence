package main

import (
	"net/http"
	"path"
)

// ServeRoot - serves the frontend root index page
func ServeRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	http.ServeFile(w, r, path.Dir("./public/index.html"))
}

// Serve 404 - served for any requests to unknown resources
func Serve404(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	http.ServeFile(w, r, path.Dir("./public/404/index.html"))
}
