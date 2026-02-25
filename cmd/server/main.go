// Package main implements a web server for the DIRELEX.
//
// The server is responsible for the following:
//   - Loading dictionary data from a gzipped JSON file.
//   - Parsing HTML templates for rendering web pages.
//   - Handling HTTP requests.
//   - Serving static assets such as CSS, JavaScript, and images.
//
// Note: Autocomplete/search functionality is implemented client-side in JavaScript.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/softcatala/direlex/internal/core"
	"github.com/softcatala/direlex/internal/server"
)

func main() {
	err := core.Init()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", server.IndexAndEntryHandler)
	mux.HandleFunc("GET /lema/{slug}", server.IndexAndEntryHandler)
	mux.HandleFunc("GET /lletra/{letter}", server.LetterHandler)
	mux.HandleFunc("GET /camp-semantic/{slug}", server.SemanticFieldHandler)
	for _, page := range core.StaticPages {
		mux.HandleFunc("GET /"+page.Path, server.BasicPageHandler(page.Path, page.Title))
	}

	mux.Handle("GET /css/", http.StripPrefix("/css/", http.FileServerFS(os.DirFS("public/css"))))
	mux.Handle("GET /js/", http.StripPrefix("/js/", http.FileServerFS(os.DirFS("public/js"))))
	mux.Handle("GET /img/", http.StripPrefix("/img/", http.FileServerFS(os.DirFS("public/img"))))
	mux.Handle("GET /favicon.svg", http.FileServerFS(os.DirFS("public")))
	mux.Handle("GET /robots.txt", http.FileServerFS(os.DirFS("public")))

	serverAddress := core.GetServerAddress()
	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: mux,
	}
	log.Println("Server started at", serverAddress)
	log.Fatal(httpServer.ListenAndServe())
}
