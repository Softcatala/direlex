module github.com/softcatala/direlex

go 1.26.0

require (
	// cmd/build-assets: JS/CSS bundling and minification
	github.com/evanw/esbuild v0.28.2
	// cmd/generate: Brotli compression
	github.com/molecule-man/go-brrr v1.3.1
	// cmd/generate: HTML minification
	github.com/tdewolff/minify/v2 v2.24.25
	// cmd/generate: Parallel execution
	golang.org/x/sync v0.25.0
)

require (
	github.com/tdewolff/parse/v2 v2.8.20 // indirect
	golang.org/x/sys v0.49.0 // indirect
)
