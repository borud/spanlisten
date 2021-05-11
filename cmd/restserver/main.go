package main

import (
	"fmt"
	"log"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>hello there %s</h1>\n", r.RemoteAddr)
	fmt.Fprint(w, "<img width=200 src=\"/static/gopher.svg\">\n")
}

func main() {
	mux := http.NewServeMux()

	// Serve static files from the filesystem
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle this with our indexhandler
	mux.HandleFunc("/", indexHandler)

	// Set up the server
	server := http.Server{
		Addr:     ":9090",
		Handler:  mux,
		ErrorLog: log.Default(),
	}

	// Start the server
	server.ListenAndServe()
}
