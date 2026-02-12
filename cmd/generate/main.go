// Package main implements the static site generator for DIRELEX.
//
// The generator is responsible for the following:
//   - Loading dictionary data from a gzipped JSON file.
//   - Parsing HTML templates for rendering web pages.
//   - Generating all static HTML pages.
//   - Minifying and compressing HTML, CSS, JS, and SVG files.
package main

import (
	"log"

	"github.com/softcatala/direlex/internal/core"
	"github.com/softcatala/direlex/internal/generator"
)

func main() {
	err := core.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = generator.GenerateStaticSite()
	if err != nil {
		log.Fatalf("Failed to generate static site: %v", err)
	}
}
