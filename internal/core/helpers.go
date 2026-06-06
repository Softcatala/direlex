package core

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"maps"
	"os"
	"slices"
	"strings"
)

// GetServerAddress returns the server address from the PORT env variable.
func GetServerAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	return ":" + port
}

// Init loads all application data and initializes templates.
// This function should be called once at startup by both the server and generator.
func Init() error {
	err := LoadDataFromFile("data/data.json.gz")
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	log.Printf("Loaded %d entries, %d semantic fields, and glossary.\n", len(AllEntries), len(SemanticFields))

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
	MainTemplate, err = template.New("main.html").Funcs(funcMap).ParseFS(templateFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return fmt.Errorf("failed to initialize templates: %w", err)
	}

	return nil
}

// LoadDataFromFile loads and processes all dictionary data from a gzipped JSON file.
// It populates the global variables: AllEntries, SemanticFields, DictionaryLetters, and Glossary.
// This function is called once at startup.
func LoadDataFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open data file %s: %w", filePath, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	var data struct {
		Entries        []Entry           `json:"entries"`
		SemanticFields []SemanticField   `json:"semantic_fields"`
		Glossary       map[string]string `json:"glossary"`
	}
	err = json.NewDecoder(gzipReader).Decode(&data)
	if err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	AllEntries = data.Entries
	SemanticFields = data.SemanticFields

	// Convert glossary strings to template.HTML to prevent escaping
	Glossary = make(map[string]template.HTML, len(data.Glossary))
	for letter, content := range data.Glossary {
		Glossary[letter] = template.HTML(content)
	}

	// Build index for fast entry lookup by slug
	entryIndexBySlug = make(map[string]int, len(AllEntries))

	// Extract unique first letters from dictionary entries for the letter browsing pages.
	// These are lowercase letters (a-z) from the normalized entry titles.
	letterMap := make(map[string]bool)
	for i, entry := range AllEntries {
		entryIndexBySlug[entry.Slug] = i
		if len(entry.NormalizedTitle) > 0 {
			firstLetter := string(entry.NormalizedTitle[0])
			letterMap[firstLetter] = true
		}
	}
	DictionaryLetters = slices.Sorted(maps.Keys(letterMap))

	return nil
}

// RenderEntry renders the HTML for a dictionary entry.
func RenderEntry(entry Entry) string {
	return fmt.Sprintf(
		`<h2>%s</h2><div>%s</div>`,
		entry.DisplayTitle,
		entry.Content,
	)
}

// RenderEntryBySlug renders the HTML for a specific entry slug.
func RenderEntryBySlug(slug string) (string, bool) {
	i, ok := entryIndexBySlug[slug]
	if !ok {
		return "", false
	}

	return RenderEntry(AllEntries[i]), true
}

// GetAdjacentEntrySlugs returns the previous and next entry slugs for a given entry slug.
// Returns empty strings for prev/next if at the beginning/end of the list.
func GetAdjacentEntrySlugs(slug string) (string, string) {
	i, ok := entryIndexBySlug[slug]
	if !ok {
		return "", ""
	}

	var prev, next string
	if i > 0 {
		prev = AllEntries[i-1].Slug
	}
	if i < len(AllEntries)-1 {
		next = AllEntries[i+1].Slug
	}

	return prev, next
}

// GetNavigationLetters returns the previous and next letters in the Catalan alphabet.
// Returns empty strings for prev/next if at the beginning/end of the alphabet.
func GetNavigationLetters(letter string) (string, string) {
	i := slices.Index(DictionaryLetters, letter)
	if i < 0 {
		return "", ""
	}

	var prev, next string
	if i > 0 {
		prev = DictionaryLetters[i-1]
	}
	if i < len(DictionaryLetters)-1 {
		next = DictionaryLetters[i+1]
	}

	return prev, next
}

// CreateHomePageData creates a fully populated PageData struct for the homepage.
func CreateHomePageData() PageData {
	return PageData{
		PlainTextTitle: "Diccionari de recursos lexicals",
		PageType:       "home",
		Letters:        DictionaryLetters,
	}
}

// CreateStaticPageData creates a fully populated PageData struct for a static page.
func CreateStaticPageData(path, title string) PageData {
	data := PageData{
		PlainTextTitle: title,
		PageType:       path,
	}

	if path == "glossari" {
		data.GlossaryLetters = slices.Sorted(maps.Keys(Glossary))
		data.GlossaryContent = Glossary
	}

	return data
}

// CreateSemanticFieldPageData creates a fully populated PageData struct for a semantic field page.
func CreateSemanticFieldPageData(title, body string) PageData {
	return PageData{
		PlainTextTitle: title,
		PageType:       "semantic-field",
		ContentHTML:    template.HTML(body),
	}
}

// CreateLetterPageData creates a fully populated PageData struct for a letter browsing page.
func CreateLetterPageData(letter string, entries []LetterEntry, prevLetter, nextLetter string) PageData {
	return PageData{
		PlainTextTitle: fmt.Sprintf("Paraules que comencen per %s", letter),
		PageType:       "letter",
		Letter:         letter,
		Entries:        entries,
		PrevLetter:     prevLetter,
		NextLetter:     nextLetter,
	}
}

// CreateEntryPageData creates a fully populated PageData struct for an entry page.
// Parameters:
//   - slug: The lema's unique identifier (e.g., "absència", "adonar-se_(de)")
//   - entryHTML: The rendered HTML content for the lema
//   - prevSlug, nextSlug: Slugs for navigation to adjacent entries
func CreateEntryPageData(slug, entryHTML, prevSlug, nextSlug string) PageData {
	return PageData{
		PlainTextTitle: strings.ReplaceAll(slug, "_", " "),
		PageType:       "entry",
		ContentHTML:    template.HTML(entryHTML),
		PrevSlug:       prevSlug,
		NextSlug:       nextSlug,
	}
}

// Create404PageData creates a fully populated PageData struct for the 404 error page.
func Create404PageData() PageData {
	return PageData{
		PlainTextTitle: "No s'ha trobat",
		PageType:       "404",
	}
}

// GetEntriesByFirstLetter returns entry data for lemes starting with the given letter.
// The letter parameter should be a single lowercase letter (a-z).
// Returns an empty slice if no entries are found for the given letter.
// Entries are assumed to be pre-sorted in Catalan locale order from the data export.
// Normalized titles are assumed to be converted in the export (lowercase, removed accents).
func GetEntriesByFirstLetter(letter string) []LetterEntry {
	var entries []LetterEntry
	for _, entry := range AllEntries {
		if len(entry.NormalizedTitle) > 0 && entry.NormalizedTitle[0] == letter[0] {
			entries = append(entries, LetterEntry{
				Slug:         entry.Slug,
				DisplayTitle: template.HTML(entry.DisplayTitle),
			})
		}
	}
	return entries
}
