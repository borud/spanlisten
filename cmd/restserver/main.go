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

	// Print query string
	fmt.Fprintf(w, "<p>Query: %s</p>", r.URL.Query())

	// Print key value pairs
	fmt.Fprint(w, "<ul>\n")
	for k, v := range r.URL.Query() {
		fmt.Fprintf(w, "<li>%s = '%s'</li>", k, v)
	}
	fmt.Fprint(w, "</ul>\n")

	// Picking out particular values
	names := r.URL.Query()["name"]

	if names != nil {
		fmt.Fprintf(w, "The names are %s", names)
	}
}

func formPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing form: %v", err), http.StatusBadRequest)
		return
	}

	// curl -X POST http://localhost:9090/login -H "Content-Type: application/x-www-form-urlencoded" -d "username=borud&password=secret"
	// curl -X POST http://localhost:9090/login -d "username=borud&password=secret"
	fmt.Fprint(w, "<h1>Form data</h1>\n")
	fmt.Fprintf(w, "Username: %s<br>\n", r.FormValue("username"))
	fmt.Fprintf(w, "Password: %s<br>\n", r.FormValue("password"))
}

func main() {
	mux := http.NewServeMux()

	// Serve static files from the filesystem
	fs := http.FileServer(http.FS(static.StaticFS))
	mux.Handle("/files/", fs)

	// Handle root with our indexHandler
	mux.HandleFunc("/", indexHandler)

	// Handle form post
	mux.HandleFunc("/login", formPostHandler)

	// Set up the server
	server := http.Server{
		Addr:     ":9090",
		Handler:  mux,
		ErrorLog: log.Default(),
	}

	// Start the server
	server.ListenAndServe()
}
