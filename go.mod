module github.com/softcatala/direlex

go 1.24.0

require (
	github.com/andybalholm/brotli v1.2.0 // cmd/generate: Brotli compression
	github.com/evanw/esbuild v0.27.2 // cmd/build-assets: JS/CSS bundling and minification
	github.com/tdewolff/minify/v2 v2.24.8 // cmd/generate: HTML minification
	golang.org/x/sync v0.19.0 // cmd/generate: Parallel execution
)

require (
	github.com/tdewolff/parse/v2 v2.8.5 // indirect
	golang.org/x/sys v0.37.0 // indirect
)
