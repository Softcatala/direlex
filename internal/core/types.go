package core

import "html/template"

// Entry represents a dictionary entry with three forms of the title:
//
// Slug: The canonical identifier used in URLs and as a unique key.
// DisplayTitle: The formatted title for display to users, may include HTML.
// NormalizedTitle: The searchable form - lowercase with accents removed.
type Entry struct {
	Slug            string `json:"title"`
	DisplayTitle    string `json:"title_display"`
	NormalizedTitle string `json:"title_normalized"`
	Content         string `json:"content"`
}

// SemanticField represents a semantic field page with a title, body content, and URL path.
type SemanticField struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Path  string `json:"path"`
}

// Represents the data for rendering a page
type PageData struct {
	// PlainTextTitle is used for rendering the page title in the template,
	// inside <title> tag, and in autocomplete input value in entry pages.
	PlainTextTitle string

	// PageType indicates the type of page being rendered
	PageType string

	// Used in index page (homepage)
	Letters []string

	// Used in entry pages
	PrevSlug string
	NextSlug string

	// Used in letter pages
	Letter     string
	PrevLetter string
	NextLetter string
	Entries    []LetterEntry

	// Used in glossary page
	GlossaryLetters []string
	GlossaryContent map[string]template.HTML

	// ContentHTML holds the main HTML content for dynamic pages
	// (entry and semantic field pages)
	ContentHTML template.HTML
}

// LetterEntry represents a minimal entry used by letter browsing pages.
type LetterEntry struct {
	Slug         string
	DisplayTitle template.HTML
}
