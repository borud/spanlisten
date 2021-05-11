package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/borud/spanlisten/pkg/static"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>hello there %s</h1>\n", r.RemoteAddr)
	fmt.Fprint(w, "<img width=200 src=\"/files/gopher.svg\">\n")
}

func main() {
	mux := http.NewServeMux()

	// Serve static files from the filesystem
	fs := http.FileServer(http.FS(static.StaticFS))
	mux.Handle("/files/", fs)

	// Handle root with our indexHandler
	mux.HandleFunc("/", indexHandler)

	log.Printf("%#v", mux)

	// Set up the server
	server := http.Server{
		Addr:     ":9090",
		Handler:  mux,
		ErrorLog: log.Default(),
	}

	// Start the server
	server.ListenAndServe()
}
