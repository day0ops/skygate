package web_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"skygate/internal/web"
)

func TestStaticIndexServed(t *testing.T) {
	sub, err := fs.Sub(web.Static, "static")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(sub))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "skygate mission control") {
		t.Fatalf("expected index.html content, got: %s", rec.Body.String())
	}
}
