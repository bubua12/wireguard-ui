package api

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func newStaticRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	bundle := strings.Repeat("console.log('wireguard');\n", 200)
	site := newStaticSite(fstest.MapFS{
		"index.html":       {Data: []byte("<!doctype html><title>app</title>")},
		"favicon.ico":      {Data: []byte{0, 0, 1, 0}},
		"assets/app-1.js":  {Data: []byte(bundle)},
		"assets/app-1.css": {Data: []byte("body{color:red}")},
	})
	r := gin.New()
	r.NoRoute(site.handle)
	return r
}

func get(r http.Handler, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestStaticSPAFallbackAndMissingAssets(t *testing.T) {
	r := newStaticRouter()

	for _, path := range []string{"/", "/peers", "/settings", "/some/deep/route"} {
		w := get(r, path, nil)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "<title>app</title>") {
			t.Fatalf("%s = %d %q", path, w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-cache" {
			t.Fatalf("%s Cache-Control = %q", path, w.Header().Get("Cache-Control"))
		}
	}

	if w := get(r, "/assets/missing.js", nil); w.Code != http.StatusNotFound {
		t.Fatalf("missing asset = %d, want 404 instead of the app shell", w.Code)
	}
	if w := get(r, "/api/nope", nil); w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `"error"`) {
		t.Fatalf("unknown api route = %d %s", w.Code, w.Body.String())
	}
	if w := get(r, "/../../etc/passwd", nil); w.Code == http.StatusOK && !strings.Contains(w.Body.String(), "<title>app</title>") {
		t.Fatalf("path traversal served something else: %q", w.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/peers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST to static route = %d", w.Code)
	}
}

func TestStaticCachingAndGzip(t *testing.T) {
	r := newStaticRouter()

	plain := get(r, "/assets/app-1.js", nil)
	if plain.Code != http.StatusOK || plain.Header().Get("Content-Encoding") != "" {
		t.Fatalf("plain = %d enc=%q", plain.Code, plain.Header().Get("Content-Encoding"))
	}
	if !strings.Contains(plain.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("hashed asset should be immutable, got %q", plain.Header().Get("Cache-Control"))
	}
	if !strings.Contains(plain.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("content type = %q", plain.Header().Get("Content-Type"))
	}

	zipped := get(r, "/assets/app-1.js", map[string]string{"Accept-Encoding": "br, gzip;q=0.8"})
	if zipped.Header().Get("Content-Encoding") != "gzip" || zipped.Body.Len() >= plain.Body.Len() {
		t.Fatalf("expected a smaller gzip body, got enc=%q %d vs %d", zipped.Header().Get("Content-Encoding"), zipped.Body.Len(), plain.Body.Len())
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	inflated, _ := io.ReadAll(zr)
	if !bytes.Equal(inflated, plain.Body.Bytes()) {
		t.Fatal("gzip body does not match the original")
	}
	if get(r, "/assets/app-1.js", map[string]string{"Accept-Encoding": "gzip;q=0"}).Header().Get("Content-Encoding") != "" {
		t.Fatal("gzip;q=0 must be honoured")
	}
	// Tiny files are not worth compressing.
	if get(r, "/assets/app-1.css", map[string]string{"Accept-Encoding": "gzip"}).Header().Get("Content-Encoding") != "" {
		t.Fatal("small file should not be gzipped")
	}

	etag := plain.Header().Get("ETag")
	if etag == "" || zipped.Header().Get("ETag") != etag {
		t.Fatalf("etag should be stable across encodings: %q vs %q", etag, zipped.Header().Get("ETag"))
	}
	if w := get(r, "/assets/app-1.js", map[string]string{"If-None-Match": etag}); w.Code != http.StatusNotModified || w.Body.Len() != 0 {
		t.Fatalf("conditional request = %d (%d bytes)", w.Code, w.Body.Len())
	}
	if w := get(r, "/assets/app-1.js", map[string]string{"If-None-Match": `"other"`}); w.Code != http.StatusOK {
		t.Fatalf("stale etag = %d", w.Code)
	}
}

func TestStaticHead(t *testing.T) {
	r := newStaticRouter()
	req := httptest.NewRequest(http.MethodHead, "/assets/app-1.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.Len() != 0 || w.Header().Get("Content-Length") == "" {
		t.Fatalf("HEAD = %d body=%d len=%q", w.Code, w.Body.Len(), w.Header().Get("Content-Length"))
	}
}
