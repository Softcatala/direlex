package server

import (
	"log"
	"net/http"

	"github.com/softcatala/direlex/internal/core"
)

// BasicPageHandler returns an HTTP handler function for rendering basic static pages.
// It takes a path and title, which are used to populate the PageData struct.
func BasicPageHandler(path, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageData := core.CreateStaticPageData(path, title)
		err := core.MainTemplate.Execute(w, pageData)
		if err != nil {
			log.Printf("Error executing template: %v", err)
		}
	}
}

// HomeHandler serves the homepage (GET /{$}).
// Client-side autocomplete is implemented in JavaScript.
// Register with the exact-path pattern "GET /{$}" so unknown paths are not
// treated as the homepage (plain "GET /" is a catch-all in net/http.ServeMux).
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	pageData := core.CreateHomePageData()
	err := core.MainTemplate.Execute(w, pageData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
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
	pageData := core.CreateEntryPageData(slug, entryHTML, prevSlug, nextSlug)
	err := core.MainTemplate.Execute(w, pageData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// NotFoundHandler serves the custom HTML 404 page for unmatched routes.
// Register last with "GET /{path...}" so more specific patterns still win.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
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
	pageData := core.CreateLetterPageData(letter, entries, prevLetter, nextLetter)
	err := core.MainTemplate.Execute(w, pageData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
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
			pageData := core.CreateSemanticFieldPageData(field.Title, field.Body)
			err := core.MainTemplate.Execute(w, pageData)
			if err != nil {
				log.Printf("Error executing template: %v", err)
			}
			return
		}
	}

	serveNotFound(w)
}

// serveNotFound renders a standard 404 Not Found error page.
func serveNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	pageData := core.Create404PageData()
	err := core.MainTemplate.Execute(w, pageData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
