// Package main implements the asset bundler for DIRELEX.
//
// This command uses esbuild to bundle and minify CSS and JavaScript files.
// It should be run before building the server or generating the static site.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

var browserTargets = []api.Engine{
	{Name: api.EngineChrome, Version: "90"},
	{Name: api.EngineFirefox, Version: "88"},
	{Name: api.EngineSafari, Version: "14"},
}

func main() {
	log.Println("Building assets...")

	err := buildCSS()
	if err != nil {
		log.Fatalf("Failed to build CSS: %v", err)
	}

	err = buildJS()
	if err != nil {
		log.Fatalf("Failed to build JS: %v", err)
	}

	log.Println("Assets built successfully!")
}

func buildCSS() error {
	log.Println("  Building CSS...")

	err := os.MkdirAll("public/css", 0o755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return buildAsset(api.BuildOptions{
		EntryPoints:       []string{"css/main.css"},
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Engines:           browserTargets,
		Outfile:           "public/css/main.min.css",
		Write:             true,
		LogLevel:          api.LogLevelInfo,
	})
}

func buildJS() error {
	log.Println("  Building JavaScript...")

	err := os.MkdirAll("public/js", 0o755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	jsFiles := []string{"search.js", "search-glossary.js"}

	for _, file := range jsFiles {
		inputPath := filepath.Join("js", file)
		base := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		outputPath := filepath.Join("public/js", base+".min.js")

		err := buildAsset(api.BuildOptions{
			EntryPoints:       []string{inputPath},
			Bundle:            true,
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			Target:            api.ES2020,
			Engines:           browserTargets,
			Format:            api.FormatIIFE,
			Outfile:           outputPath,
			Write:             true,
			LogLevel:          api.LogLevelInfo,
		})
		if err != nil {
			return fmt.Errorf("failed to build %s: %w", file, err)
		}
	}

	return nil
}

func buildAsset(options api.BuildOptions) error {
	result := api.Build(options)
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			log.Printf("Build error: %s", err.Text)
		}
		return fmt.Errorf("build failed with %d errors", len(result.Errors))
	}

	return nil
}
