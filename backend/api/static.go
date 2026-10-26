package api

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const gzipMinSize = 1024

// staticEntry is one embedded frontend file, ready to serve.
type staticEntry struct {
	body  []byte
	gz    []byte // nil when compression is not worth it
	etag  string
	ctype string
}

// staticSite serves the embedded single page app: long-lived caching for the
// hashed bundle files, revalidation for everything else, gzip for text assets,
// and index.html for unknown routes so client-side navigation survives reloads.
type staticSite struct {
	fsys  fs.FS
	mu    sync.RWMutex
	cache map[string]*staticEntry
}

func newStaticSite(fsys fs.FS) *staticSite {
	return &staticSite{fsys: fsys, cache: make(map[string]*staticEntry)}
}

func (s *staticSite) entry(name string) (*staticEntry, bool) {
	s.mu.RLock()
	e, ok := s.cache[name]
	s.mu.RUnlock()
	if ok {
		return e, true
	}

	body, err := fs.ReadFile(s.fsys, name)
	if err != nil {
		return nil, false
	}
	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = http.DetectContentType(body)
	}
	sum := sha256.Sum256(body)
	e = &staticEntry{body: body, ctype: ctype, etag: `W/"` + hex.EncodeToString(sum[:8]) + `"`}
	if len(body) >= gzipMinSize && compressible(ctype) {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		_, _ = zw.Write(body)
		if zw.Close() == nil && buf.Len() < len(body) {
			e.gz = buf.Bytes()
		}
	}

	s.mu.Lock()
	s.cache[name] = e
	s.mu.Unlock()
	return e, true
}

func compressible(ctype string) bool {
	return strings.HasPrefix(ctype, "text/") ||
		strings.Contains(ctype, "javascript") ||
		strings.Contains(ctype, "json") ||
		strings.Contains(ctype, "svg") ||
		strings.Contains(ctype, "xml")
}

func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		fields := strings.Split(strings.TrimSpace(part), ";")
		if strings.EqualFold(strings.TrimSpace(fields[0]), "gzip") {
			for _, param := range fields[1:] {
				if strings.TrimSpace(param) == "q=0" {
					return false
				}
			}
			return true
		}
	}
	return false
}

func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || strings.TrimPrefix(candidate, "W/") == strings.TrimPrefix(etag, "W/") {
			return true
		}
	}
	return false
}

func (s *staticSite) handle(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(path.Clean("/"+c.Request.URL.Path), "/")
	if name == "" {
		name = "index.html"
	}
	e, ok := s.entry(name)
	if !ok {
		// A missing script or stylesheet must stay a 404; answering it with the
		// app shell would only surface later as a confusing syntax error.
		if path.Ext(name) != "" {
			c.String(http.StatusNotFound, "not found")
			return
		}
		name = "index.html"
		if e, ok = s.entry(name); !ok {
			c.String(http.StatusNotFound, "not found")
			return
		}
	}

	h := c.Writer.Header()
	h.Set("Content-Type", e.ctype)
	h.Set("ETag", e.etag)
	h.Set("Vary", "Accept-Encoding")
	if strings.HasPrefix(name, "assets/") {
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		h.Set("Cache-Control", "no-cache")
	}
	if match := c.GetHeader("If-None-Match"); match != "" && etagMatches(match, e.etag) {
		c.Status(http.StatusNotModified)
		return
	}

	body := e.body
	if e.gz != nil && acceptsGzip(c.GetHeader("Accept-Encoding")) {
		body = e.gz
		h.Set("Content-Encoding", "gzip")
	}
	h.Set("Content-Length", strconv.Itoa(len(body)))
	c.Status(http.StatusOK)
	if c.Request.Method == http.MethodGet {
		_, _ = c.Writer.Write(body)
	}
}
