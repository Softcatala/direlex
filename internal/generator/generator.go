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
	"sync/atomic"

	"github.com/andybalholm/brotli"
	"github.com/softcatala/direlex/internal/core"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"golang.org/x/sync/errgroup"
)

const (
	OutputDir = "build"
)

// GenerateStaticSite generates all static HTML files for the dictionary website.
func GenerateStaticSite() error {
	log.Println("Starting static site generation...")

	err := os.RemoveAll(OutputDir)
	if err != nil {
		return fmt.Errorf("failed to remove old output directory: %w", err)
	}

	err = os.MkdirAll(OutputDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	log.Println("Generating homepage...")
	err = generateHomePage()
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
	err = generate404Page()
	if err != nil {
		return fmt.Errorf("failed to generate 404 page: %w", err)
	}

	log.Println("Copying assets...")
	err = os.CopyFS(OutputDir, os.DirFS("public"))
	if err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	log.Println("Compressing files...")
	err = compressFiles()
	if err != nil {
		return fmt.Errorf("failed to compress files: %w", err)
	}

	log.Println("Static site generation completed successfully.")
	log.Printf("Output directory: %s\n", OutputDir)
	return nil
}

// generateHomePage generates the homepage (index.html).
func generateHomePage() error {
	pageData := core.CreateHomePageData()
	return writeHTMLFile("index.html", pageData)
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
	pageData := core.CreateEntryPageData(entry.Slug, entryHTML, prevSlug, nextSlug)
	outputPath := filepath.Join("lema", entry.Slug+".html")

	err := writeHTMLFile(outputPath, pageData)
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
		pageData := core.CreateLetterPageData(letter, entries, prevLetter, nextLetter)

		outputPath := filepath.Join("lletra", letter+".html")
		err := writeHTMLFile(outputPath, pageData)
		if err != nil {
			return fmt.Errorf("failed to generate letter page %s: %w", letter, err)
		}
	}

	return nil
}

// generateStaticPages generates static pages as flat files.
func generateStaticPages() error {
	for _, page := range core.StaticPages {
		pageData := core.CreateStaticPageData(page.Path, page.Title)

		outputPath := page.Path + ".html"
		err := writeHTMLFile(outputPath, pageData)
		if err != nil {
			return fmt.Errorf("failed to generate page %s: %w", page.Path, err)
		}
	}

	return nil
}

// generateSemanticFieldPages generates all semantic field pages as flat files.
func generateSemanticFieldPages() error {
	for _, field := range core.SemanticFields {
		pageData := core.CreateSemanticFieldPageData(field.Title, field.Body)

		outputPath := filepath.Join("camp-semantic", field.Path+".html")
		err := writeHTMLFile(outputPath, pageData)
		if err != nil {
			return fmt.Errorf("failed to generate semantic field page %s: %w", field.Path, err)
		}
	}

	return nil
}

// generate404Page generates the 404 error page.
func generate404Page() error {
	pageData := core.Create404PageData()
	return writeHTMLFile("404.html", pageData)
}

// writeHTMLFile writes a rendered HTML page to the output directory.
func writeHTMLFile(relativePath string, data core.PageData) error {
	fullPath := filepath.Join(OutputDir, relativePath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = core.MainTemplate.Execute(&buf, data)
	if err != nil {
		return err
	}

	m := minify.New()
	htmlMinifier := &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
	}
	m.AddFunc("text/html", htmlMinifier.Minify)
	minifiedBytes, err := m.Bytes("text/html", buf.Bytes())
	if err != nil {
		log.Printf("warning: could not minify %s: %v. Original content will be used.", relativePath, err)
		minifiedBytes = buf.Bytes()
	}

	err = os.WriteFile(fullPath, minifiedBytes, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// compressFiles compresses files using GZIP and Brotli in parallel.
func compressFiles() error {
	filesToCompress, err := getFilesToCompress()
	if err != nil {
		return err
	}

	total := len(filesToCompress)
	g := new(errgroup.Group)
	g.SetLimit(runtime.NumCPU())
	var processed atomic.Int64

	for _, path := range filesToCompress {
		g.Go(func() error {
			err := compressFile(path)
			if err != nil {
				return err
			}

			n := int(processed.Add(1))
			if n%100 == 0 || n == total {
				log.Printf("  Compressed %d/%d files\n", n, total)
			}

			return nil
		})
	}

	return g.Wait()
}

// getFilesToCompress returns a list of files that should be compressed.
func getFilesToCompress() ([]string, error) {
	var files []string
	err := filepath.WalkDir(OutputDir, func(path string, d fs.DirEntry, err error) error {
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
// Uses maximum compression level (9) for best compression ratio.
func compressFileGzip(filePath string, content []byte) error {
	gzPath := filePath + ".gz"
	file, err := os.Create(gzPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	gzWriter, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}
	defer gzWriter.Close()

	_, err = gzWriter.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}

// compressFileBrotli creates a pre-compressed .br version of the given file.
// Uses maximum compression level (11) for best compression ratio.
func compressFileBrotli(filePath string, content []byte) error {
	brPath := filePath + ".br"
	file, err := os.Create(brPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	brWriter := brotli.NewWriterLevel(file, brotli.BestCompression)
	defer brWriter.Close()

	_, err = brWriter.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}
