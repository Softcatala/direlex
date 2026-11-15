// Package generator builds the dictionary's static website.
package generator

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	brotli "github.com/molecule-man/go-brrr"
	"github.com/softcatala/direlex/internal/core"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"golang.org/x/sync/errgroup"
)

const outputDir = "build"

var htmlMinifier = newHTMLMinifier()

// buildDate is shared by all generated pages so their asset URLs are invalidated together.
var buildDate = time.Now().UTC().Format("20060102-150405")

// newHTMLMinifier returns a minify instance configured for HTML pages.
func newHTMLMinifier() *minify.M {
	m := minify.New()
	m.AddFunc("text/html", (&html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
	}).Minify)
	return m
}

// GenerateStaticSite generates all static HTML files for the dictionary website.
func GenerateStaticSite() error {
	log.Println("Starting static site generation...")

	err := os.RemoveAll(outputDir)
	if err != nil {
		return fmt.Errorf("failed to remove old output directory: %w", err)
	}

	err = os.MkdirAll(outputDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	log.Println("Generating homepage...")
	err = writeHTMLFile("index.html", core.CreateHomePageData())
	if err != nil {
		return fmt.Errorf("failed to generate homepage: %w", err)
	}

	log.Printf("Generating %d entry pages...\n", len(core.AllEntries))
	err = generateEntryPages()
	if err != nil {
		return fmt.Errorf("failed to generate entry pages: %w", err)
	}

	log.Printf("Generating %d letter pages...\n", len(core.DictionaryLetters))
	err = generateLetterPages()
	if err != nil {
		return fmt.Errorf("failed to generate letter pages: %w", err)
	}

	log.Println("Generating static pages...")
	err = generateStaticPages()
	if err != nil {
		return fmt.Errorf("failed to generate static pages: %w", err)
	}

	log.Printf("Generating %d semantic field pages...\n", len(core.SemanticFields))
	err = generateSemanticFieldPages()
	if err != nil {
		return fmt.Errorf("failed to generate semantic field pages: %w", err)
	}

	log.Println("Generating 404 page...")
	err = writeHTMLFile("404.html", core.Create404PageData())
	if err != nil {
		return fmt.Errorf("failed to generate 404 page: %w", err)
	}

	log.Println("Copying assets...")
	err = os.CopyFS(outputDir, os.DirFS("public"))
	if err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	log.Println("Compressing files...")
	err = compressFiles()
	if err != nil {
		return fmt.Errorf("failed to compress files: %w", err)
	}

	log.Println("Static site generation completed successfully.")
	log.Printf("Output directory: %s\n", outputDir)
	return nil
}

// generateEntryPages generates all individual entry pages.
func generateEntryPages() error {
	for _, entry := range core.AllEntries {
		err := generateEntryPage(entry)
		if err != nil {
			return err
		}
	}

	return nil
}

// generateEntryPage generates a single dictionary entry page.
func generateEntryPage(entry core.Entry) error {
	entryHTML := core.RenderEntry(entry)
	prevSlug, nextSlug := core.GetAdjacentEntrySlugs(entry.Slug)
	outputPath := filepath.Join("lema", entry.Slug+".html")

	err := writeHTMLFile(
		outputPath,
		core.CreateEntryPageData(entry.Slug, entryHTML, prevSlug, nextSlug),
	)
	if err != nil {
		return fmt.Errorf("failed to generate entry %s: %w", entry.Slug, err)
	}

	return nil
}

// generateLetterPages generates all letter browsing pages as flat files.
func generateLetterPages() error {
	for _, letter := range core.DictionaryLetters {
		entries := core.GetEntriesByFirstLetter(letter)
		if len(entries) == 0 {
			continue
		}

		prevLetter, nextLetter := core.GetNavigationLetters(letter)

		outputPath := filepath.Join("lletra", letter+".html")
		err := writeHTMLFile(outputPath, core.CreateLetterPageData(letter, entries, prevLetter, nextLetter))
		if err != nil {
			return fmt.Errorf("failed to generate letter page %s: %w", letter, err)
		}
	}

	return nil
}

// generateStaticPages generates static pages as flat files.
func generateStaticPages() error {
	for _, page := range core.StaticPages {
		outputPath := page.Path + ".html"
		err := writeHTMLFile(outputPath, core.CreateStaticPageData(page.Path, page.Title))
		if err != nil {
			return fmt.Errorf("failed to generate page %s: %w", page.Path, err)
		}
	}

	return nil
}

// generateSemanticFieldPages generates all semantic field pages as flat files.
func generateSemanticFieldPages() error {
	for _, field := range core.SemanticFields {
		outputPath := filepath.Join("camp-semantic", field.Path+".html")
		err := writeHTMLFile(outputPath, core.CreateSemanticFieldPageData(field.Title, field.Body))
		if err != nil {
			return fmt.Errorf("failed to generate semantic field page %s: %w", field.Path, err)
		}
	}

	return nil
}

// writeHTMLFile writes a rendered HTML page to the output directory.
func writeHTMLFile(relativePath string, data core.PageData) error {
	data.AssetVersion = buildDate
	fullPath := filepath.Join(outputDir, relativePath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = core.MainTemplate.Execute(&buf, data)
	if err != nil {
		return err
	}

	minifiedBytes, err := htmlMinifier.Bytes("text/html", buf.Bytes())
	if err != nil {
		log.Printf("warning: could not minify %s: %v. Original content will be used.", relativePath, err)
		minifiedBytes = buf.Bytes()
	}

	return os.WriteFile(fullPath, minifiedBytes, 0o644)
}

// compressFiles compresses files using GZIP and Brotli in parallel.
func compressFiles() error {
	filesToCompress, err := getFilesToCompress()
	if err != nil {
		return err
	}

	var g errgroup.Group
	g.SetLimit(runtime.GOMAXPROCS(0))

	for _, path := range filesToCompress {
		g.Go(func() error {
			return compressFile(path)
		})
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	log.Printf("  Compressed %d files\n", len(filesToCompress))
	return nil
}

// getFilesToCompress returns a list of files that should be compressed.
func getFilesToCompress() ([]string, error) {
	var files []string
	err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && shouldCompress(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// compressFile reads a file and creates its .gz and .br versions.
func compressFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	err = compressFileGzip(path, content)
	if err != nil {
		return fmt.Errorf("failed to gzip %s: %w", path, err)
	}

	err = compressFileBrotli(path, content)
	if err != nil {
		return fmt.Errorf("failed to brotli %s: %w", path, err)
	}

	return nil
}

// shouldCompress returns true if the file should be compressed based on its extension.
func shouldCompress(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".css", ".js", ".svg", ".html":
		return true
	default:
		return false
	}
}

// compressFileGzip creates a pre-compressed .gz version of the given file.
func compressFileGzip(filePath string, content []byte) error {
	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}

	_, err = gzWriter.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	err = gzWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return os.WriteFile(filePath+".gz", buf.Bytes(), 0o644)
}

// compressFileBrotli creates a pre-compressed .br version of the given file.
func compressFileBrotli(filePath string, content []byte) error {
	compressed, err := brotli.Compress(content, brotli.BestCompression)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath+".br", compressed, 0o644)
}
