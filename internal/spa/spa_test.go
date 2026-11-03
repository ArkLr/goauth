package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<div id=root>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "assets"), 0o755)
	os.WriteFile(filepath.Join(dir, "assets", "app-abc.js"), []byte("console.log(1)"), 0o644)

	h := Handler(dir)

	tests := []struct {
		path, wantBody, wantCache string
	}{
		{"/", "<div id=root>", "no-cache"},
		{"/posts/3/edit", "<div id=root>", "no-cache"},      // client route -> index.html
		{"/assets/missing.js", "<div id=root>", "no-cache"}, // unknown file -> index.html
		{"/assets/app-abc.js", "console.log(1)", "immutable"},
		{"/../../etc/passwd", "<div id=root>", "no-cache"}, // no traversal
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", tt.path, rec.Code)
		}
		if got := rec.Body.String(); got != tt.wantBody {
			t.Errorf("%s: body %q, want %q", tt.path, got, tt.wantBody)
		}
		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, tt.wantCache) {
			t.Errorf("%s: Cache-Control %q, want %q", tt.path, cc, tt.wantCache)
		}
	}
}

func TestHandlerNotBuilt(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(t.TempDir()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status %d, want 503", rec.Code)
	}
}
