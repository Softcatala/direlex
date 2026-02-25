package core

import (
	"embed"
	"html/template"
)

// AllEntries contains all dictionary entries loaded from the data file.
var AllEntries []Entry

// entryIndexBySlug maps an entry slug to its index in AllEntries.
// It is built in LoadDataFromFile and treated as read-only afterwards.
var entryIndexBySlug map[string]int

// DictionaryLetters contains the alphabet lowercase letters used at the start of a word.
// It is populated dynamically from the entries.
var DictionaryLetters []string

// Glossary contains the raw glossary data loaded from the data file.
// It maps uppercase letters to HTML content for that letter's content.
var Glossary map[string]template.HTML

// SemanticFields contains all semantic field pages loaded from the data file.
var SemanticFields []SemanticField

// StaticPages contains the registry of static pages in the application.
var StaticPages = []struct {
	Path  string
	Title string
}{
	{"sobre-el-direlex", "Sobre el DIRELEX"},
	{"instruccions", "Instruccions d'ús"},
	{"abreviatures", "Abreviatures"},
	{"glossari", "Glossari"},
	{"credits", "Crèdits"},
}

// MainTemplate is the parsed HTML template.
var MainTemplate *template.Template

//go:embed templates/*
var templateFS embed.FS
