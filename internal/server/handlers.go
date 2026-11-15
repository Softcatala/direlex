// Package server contains the HTTP handlers for the dictionary website.
package server

import (
	"log"
	"net/http"

	"github.com/softcatala/direlex/internal/core"
)

// BasicPageHandler returns an HTTP handler function for rendering basic static pages.
// It takes a path and title, which are used to populate the PageData struct.
func BasicPageHandler(path, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		renderPage(w, core.CreateStaticPageData(path, title))
	}
}

// HomeHandler serves the homepage (GET /{$}).
// Client-side autocomplete is implemented in JavaScript.
// Register with the exact-path pattern "GET /{$}" so unknown paths are not
// treated as the homepage (plain "GET /" is a catch-all in net/http.ServeMux).
func HomeHandler(w http.ResponseWriter, _ *http.Request) {
	renderPage(w, core.CreateHomePageData())
}

// EntryHandler serves an individual dictionary entry (GET /lema/{slug}).
// Client-side autocomplete is implemented in JavaScript.
// Serves a 404 page when the slug does not match any entry.
func EntryHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	entryHTML, ok := core.RenderEntryBySlug(slug)
	if !ok {
		serveNotFound(w)
		return
	}

	prevSlug, nextSlug := core.GetAdjacentEntrySlugs(slug)
	renderPage(w, core.CreateEntryPageData(slug, entryHTML, prevSlug, nextSlug))
}

// NotFoundHandler serves the custom HTML 404 page for unmatched routes.
// Register last with "GET /{path...}" so more specific patterns still win.
func NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	serveNotFound(w)
}

// LetterHandler handles requests for browsing dictionary entries by the first letter.
// It expects a URL path in the format /lletra/{letter}, where {letter} is a single lowercase letter (a-z).
// If the letter is valid and has lemes that start with it, it renders a page with a list of links (to /lema/{slug}).
//
// Additionally:
//   - Serves a 404 page for invalid letters or letters with no entries.
//   - Does not sort lemes, as this should be sorted using the Catalan locale on export time.
func LetterHandler(w http.ResponseWriter, r *http.Request) {
	letter := r.PathValue("letter")
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		serveNotFound(w)
		return
	}

	entries := core.GetEntriesByFirstLetter(letter)
	if len(entries) == 0 {
		serveNotFound(w)
		return
	}

	prevLetter, nextLetter := core.GetNavigationLetters(letter)
	renderPage(w, core.CreateLetterPageData(letter, entries, prevLetter, nextLetter))
}

// SemanticFieldHandler handles requests for semantic field pages.
// It expects a URL path in the format /camp-semantic/{slug}.
// If the slug matches a semantic field, it renders the page with the title and body.
//
// Additionally:
//   - Serves a 404 page for non-existent semantic fields.
func SemanticFieldHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	for _, field := range core.SemanticFields {
		if field.Path == slug {
			renderPage(w, core.CreateSemanticFieldPageData(field.Title, field.Body))
			return
		}
	}

	serveNotFound(w)
}

// renderPage executes the main template and logs any rendering error.
func renderPage(w http.ResponseWriter, pageData core.PageData) {
	err := core.MainTemplate.Execute(w, pageData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// serveNotFound renders a standard 404 Not Found error page.
func serveNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	renderPage(w, core.Create404PageData())
}
