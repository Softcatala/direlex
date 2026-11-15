package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/softcatala/direlex/internal/core"
	"github.com/softcatala/direlex/internal/server"
)

func TestRoutingAndNotFound(t *testing.T) {
	// data/data.json.gz is resolved relative to the process working directory.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../.."))
	err := os.Chdir(root)
	if err != nil {
		t.Fatal(err)
	}

	err = core.Init()
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", server.HomeHandler)
	mux.HandleFunc("GET /lema/{slug}", server.EntryHandler)
	mux.HandleFunc("GET /lletra/{letter}", server.LetterHandler)
	mux.HandleFunc("GET /camp-semantic/{slug}", server.SemanticFieldHandler)
	for _, page := range core.StaticPages {
		mux.HandleFunc("GET /"+page.Path, server.BasicPageHandler(page.Path, page.Title))
	}
	mux.HandleFunc("GET /{path...}", server.NotFoundHandler)

	checks := []struct {
		path       string
		wantStatus int
		wantSubstr string
	}{
		{"/", 200, "Llista de lemes"},
		{"/aserwerwe", 404, "No s'ha trobat"},
		{"/lema/abril", 200, "abril"},
		{"/lema/does-not-exist-xyz", 404, "No s'ha trobat"},
		{"/sobre-el-direlex", 200, "Sobre"},
		{"/foo/bar", 404, "No s'ha trobat"},
	}

	for _, c := range checks {
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			body, _ := io.ReadAll(rr.Result().Body)
			if rr.Code != c.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, c.wantStatus)
			}
			if !strings.Contains(string(body), c.wantSubstr) {
				t.Fatalf("body missing %q; got prefix: %.200q", c.wantSubstr, body)
			}
		})
	}
}
