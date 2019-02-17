package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	c, db := getConfig()
	initLogger(c.LogLevel)
	initDatabase(db)

	r := mux.NewRouter()
	// Serve the main page
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.ListenAndServe(":8000", r)
}
