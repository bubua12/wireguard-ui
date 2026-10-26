package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wireguard-ui/config"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

// testIface is an interface name no machine is going to have, so handlers never
// try to touch a live WireGuard device while the tests run.
const testIface = "wgtest9"

// newTestRouter wires a router to a throwaway database and config directory and
// puts a stub `wg` binary on PATH that only knows how to make keys.
func newTestRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()

	bin := t.TempDir()
	script := "#!/bin/sh\n" +
		"case \"$1\" in\n" +
		"  genkey|genpsk) head -c 32 /dev/urandom | base64 ;;\n" +
		"  pubkey) read -r priv; echo \"pub-$priv\" ;;\n" +
		"  *) echo \"stub wg: unsupported: $*\" >&2; exit 1 ;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(bin, "wg"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	confDir := t.TempDir()
	wg.SetConfigDir(confDir)
	if err := db.Init(filepath.Join(t.TempDir(), "t.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.CreateUser("admin", "correct-horse"); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{JWTSecret: []byte("unit-test-secret-0123456789"), CORSOrigins: []string{"http://localhost:5173"}}
	return SetupRouter(cfg), confDir
}

func call(t *testing.T, r http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not a JSON object: %v\n%s", err, w.Body.String())
	}
	return out
}

func login(t *testing.T, r http.Handler, password string) string {
	t.Helper()
	w := call(t, r, http.MethodPost, "/api/login", "", map[string]string{"username": "admin", "password": password})
	if w.Code != http.StatusOK {
		t.Fatalf("login = %d %s", w.Code, w.Body.String())
	}
	token, _ := decode(t, w)["token"].(string)
	if token == "" {
		t.Fatal("login returned no token")
	}
	return token
}

func TestChangePasswordSignsOtherSessionsOut(t *testing.T) {
	r, _ := newTestRouter(t)

	if w := call(t, r, http.MethodGet, "/api/status", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d", w.Code)
	}

	old := login(t, r, "correct-horse")
	if w := call(t, r, http.MethodGet, "/api/status", old, nil); w.Code != http.StatusOK {
		t.Fatalf("status with token = %d %s", w.Code, w.Body.String())
	}

	other := login(t, r, "correct-horse")
	w := call(t, r, http.MethodPost, "/api/change-password", old, map[string]string{
		"old_password": "correct-horse", "new_password": "battery-staple",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("change-password = %d %s", w.Code, w.Body.String())
	}
	fresh, _ := decode(t, w)["token"].(string)
	if fresh == "" {
		t.Fatal("change-password should return a fresh token for the current session")
	}

	for name, tok := range map[string]string{"changing session's old token": old, "other session": other} {
		if w := call(t, r, http.MethodGet, "/api/status", tok, nil); w.Code != http.StatusUnauthorized {
			t.Fatalf("%s should be signed out, got %d", name, w.Code)
		}
	}
	if w := call(t, r, http.MethodGet, "/api/status", fresh, nil); w.Code != http.StatusOK {
		t.Fatalf("fresh token rejected: %d %s", w.Code, w.Body.String())
	}
}

func TestTokenWithoutExpiryOrWrongKeyRejected(t *testing.T) {
	r, _ := newTestRouter(t)
	for _, tok := range []string{"", "garbage", "a.b.c"} {
		if w := call(t, r, http.MethodGet, "/api/peers", tok, nil); w.Code != http.StatusUnauthorized {
			t.Fatalf("token %q = %d", tok, w.Code)
		}
	}
}

func createServer(t *testing.T, r http.Handler, token string) {
	t.Helper()
	w := call(t, r, http.MethodPost, "/api/server", token, map[string]any{
		"name": "Test VPN", "interface": testIface, "address": "10.0.0.1/24",
		"endpoint": "vpn.example.com:51820", "dns": "1.1.1.1", "mtu": 1420,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create server = %d %s", w.Code, w.Body.String())
	}
}

func TestPeerLifecycle(t *testing.T) {
	r, confDir := newTestRouter(t)
	token := login(t, r, "correct-horse")
	createServer(t, r, token)

	// Auto allocation skips the server address.
	w := call(t, r, http.MethodGet, "/api/peers/next-ip", token, nil)
	if got := decode(t, w)["ip"]; got != "10.0.0.2/32" {
		t.Fatalf("next-ip = %v", got)
	}

	w = call(t, r, http.MethodPost, "/api/peers", token, map[string]string{"name": "  laptop  "})
	if w.Code != http.StatusOK {
		t.Fatalf("create peer = %d %s", w.Code, w.Body.String())
	}
	first := decode(t, w)
	if first["name"] != "laptop" || first["allowed_ips"] != "10.0.0.2/32" || first["has_private_key"] != true {
		t.Fatalf("unexpected peer: %v", first)
	}
	if _, leaked := first["private_key"]; leaked {
		t.Fatal("private key must never be serialised")
	}
	if warn, ok := first["warning"]; ok {
		t.Fatalf("persisting the config should not warn: %v", warn)
	}
	firstID := int(first["id"].(float64))

	w = call(t, r, http.MethodPost, "/api/peers", token, map[string]string{"name": "phone", "allowed_ips": "10.0.0.20/32"})
	if w.Code != http.StatusOK {
		t.Fatalf("create phone = %d %s", w.Code, w.Body.String())
	}
	secondID := int(decode(t, w)["id"].(float64))

	// Validation: bad names, duplicate and out-of-range addresses.
	for _, body := range []map[string]string{
		{"name": ""},
		{"name": strings.Repeat("x", 65)},
		{"name": "dup", "allowed_ips": "10.0.0.20/32"},
		{"name": "outside", "allowed_ips": "192.168.1.5/32"},
		{"name": "server", "allowed_ips": "10.0.0.1/32"},
	} {
		if w := call(t, r, http.MethodPost, "/api/peers", token, body); w.Code != http.StatusBadRequest {
			t.Fatalf("create %v = %d, want 400", body, w.Code)
		}
	}

	// Edit name, keepalive and address in one go.
	path := "/api/peers/" + itoa(firstID)
	w = call(t, r, http.MethodPut, path, token, map[string]any{
		"name": "work laptop", "persistent_keepalive": 0, "allowed_ips": "10.0.0.30/32",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update = %d %s", w.Code, w.Body.String())
	}
	upd := decode(t, w)
	if upd["name"] != "work laptop" || upd["allowed_ips"] != "10.0.0.30/32" || upd["persistent_keepalive"] != float64(0) {
		t.Fatalf("updated peer = %v", upd)
	}

	// Taking another peer's address, or an invalid value, is refused and changes nothing.
	for _, body := range []map[string]any{
		{"allowed_ips": "10.0.0.20/32"},
		{"allowed_ips": "10.0.0.0/24"},
		{"name": "   "},
		{"persistent_keepalive": 70000},
	} {
		if w := call(t, r, http.MethodPut, path, token, body); w.Code != http.StatusBadRequest {
			t.Fatalf("update %v = %d, want 400 (%s)", body, w.Code, w.Body.String())
		}
	}
	peers := decodeList(t, call(t, r, http.MethodGet, "/api/peers", token, nil))
	if len(peers) != 2 {
		t.Fatalf("peers = %v", peers)
	}

	// Config file only lists enabled peers and reflects the new address.
	conf, err := os.ReadFile(filepath.Join(confDir, testIface+".conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(conf), "AllowedIPs = 10.0.0.30/32") || strings.Contains(string(conf), "10.0.0.2/32") {
		t.Fatalf("server config out of date:\n%s", conf)
	}

	// Disabling drops the peer from the file.
	if w := call(t, r, http.MethodPost, path+"/toggle", token, map[string]bool{"enabled": false}); w.Code != http.StatusOK {
		t.Fatalf("toggle = %d %s", w.Code, w.Body.String())
	}
	conf, _ = os.ReadFile(filepath.Join(confDir, testIface+".conf"))
	if strings.Contains(string(conf), "10.0.0.30/32") {
		t.Fatalf("disabled peer still in config:\n%s", conf)
	}

	// Client config download.
	w = call(t, r, http.MethodGet, path+"/config", token, nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Address = 10.0.0.30/32") {
		t.Fatalf("config = %d\n%s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("client config must not be cached")
	}

	if w := call(t, r, http.MethodDelete, "/api/peers/"+itoa(secondID), token, nil); w.Code != http.StatusOK {
		t.Fatalf("delete = %d %s", w.Code, w.Body.String())
	}
	if peers := decodeList(t, call(t, r, http.MethodGet, "/api/peers", token, nil)); len(peers) != 1 {
		t.Fatalf("after delete: %v", peers)
	}
}

func TestUpdateServerValidationAndClearing(t *testing.T) {
	r, _ := newTestRouter(t)
	token := login(t, r, "correct-horse")
	createServer(t, r, token)

	for _, body := range []map[string]any{
		{"dns": "8.8.8.8\nPostUp = id"},
		{"dns": "not a dns"},
		{"mtu": 100},
		{"endpoint": "no-port"},
		{"endpoint": "bad host:51820"},
		{"listen_port": 0},
		{"interface": "../etc"},
	} {
		if w := call(t, r, http.MethodPut, "/api/server", token, body); w.Code != http.StatusBadRequest {
			t.Fatalf("update %v = %d, want 400 (%s)", body, w.Code, w.Body.String())
		}
	}

	// DNS and MTU can be cleared; a normalised list is stored.
	w := call(t, r, http.MethodPut, "/api/server", token, map[string]any{"dns": "", "mtu": 0})
	if w.Code != http.StatusOK {
		t.Fatalf("clear = %d %s", w.Code, w.Body.String())
	}
	got := decode(t, w)
	if got["dns"] != "" || got["mtu"] != float64(0) || got["restart_required"] != false {
		t.Fatalf("after clearing: %v", got)
	}
	w = call(t, r, http.MethodPut, "/api/server", token, map[string]any{"dns": " 1.1.1.1 ,8.8.8.8"})
	if got := decode(t, w); got["dns"] != "1.1.1.1, 8.8.8.8" {
		t.Fatalf("dns = %v", got["dns"])
	}

	// Omitted fields stay untouched.
	if got := decode(t, call(t, r, http.MethodGet, "/api/server", token, nil)); got["endpoint"] != "vpn.example.com:51820" || got["interface"] != testIface {
		t.Fatalf("untouched fields changed: %v", got)
	}
}

func TestStatusAndTrafficWithInterfaceDown(t *testing.T) {
	r, _ := newTestRouter(t)
	token := login(t, r, "correct-horse")

	st := decode(t, call(t, r, http.MethodGet, "/api/status", token, nil))
	if st["configured"] != false {
		t.Fatalf("unconfigured status = %v", st)
	}

	createServer(t, r, token)
	call(t, r, http.MethodPost, "/api/peers", token, map[string]string{"name": "a"})

	st = decode(t, call(t, r, http.MethodGet, "/api/status", token, nil))
	if st["configured"] != true || st["interface_up"] != false || st["peer_count"] != float64(1) || st["online_count"] != float64(0) {
		t.Fatalf("status = %v", st)
	}
	if st["public_key"] == "" || st["restart_required"] != false {
		t.Fatalf("status = %v", st)
	}

	ps := decodeList(t, call(t, r, http.MethodGet, "/api/peers/status", token, nil))
	// With the interface down nothing is known about handshakes: null, not "never".
	if len(ps) != 1 || ps[0]["online"] != false || ps[0]["handshake_age"] != nil {
		t.Fatalf("peer status = %v", ps)
	}

	tr := decode(t, call(t, r, http.MethodGet, "/api/traffic?days=7", token, nil))
	if days, _ := tr["days"].([]any); len(days) != 7 {
		t.Fatalf("traffic days = %v", tr["days"])
	}
	if w := call(t, r, http.MethodGet, "/api/traffic?days=abc", token, nil); w.Code != http.StatusBadRequest {
		t.Fatalf("bad days = %d", w.Code)
	}
}

func TestInterfaceControlRejectsBadRequests(t *testing.T) {
	r, _ := newTestRouter(t)
	token := login(t, r, "correct-horse")

	if w := call(t, r, http.MethodPost, "/api/interface/up", token, nil); w.Code != http.StatusNotFound {
		t.Fatalf("no server = %d", w.Code)
	}
	createServer(t, r, token)
	if w := call(t, r, http.MethodPost, "/api/interface/explode", token, nil); w.Code != http.StatusNotFound {
		t.Fatalf("unknown action = %d", w.Code)
	}
	// The interface is not running, so stopping it is a conflict, not a crash.
	if w := call(t, r, http.MethodPost, "/api/interface/down", token, nil); w.Code != http.StatusConflict {
		t.Fatalf("down while down = %d %s", w.Code, w.Body.String())
	}
	if w := call(t, r, http.MethodPost, "/api/interface/up", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated up = %d", w.Code)
	}
}

func TestSecurityHeadersAndBodyLimit(t *testing.T) {
	r, _ := newTestRouter(t)

	w := call(t, r, http.MethodGet, "/api/health", "", nil)
	if w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("missing security headers: %v", w.Header())
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'self'") {
		t.Fatalf("missing CSP: %v", w.Header())
	}

	big := map[string]string{"username": "admin", "password": strings.Repeat("x", 2<<20)}
	if w := call(t, r, http.MethodPost, "/api/login", "", big); w.Code != http.StatusBadRequest {
		t.Fatalf("oversized body = %d", w.Code)
	}
}

func decodeList(t *testing.T, w *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not a JSON array: %v\n%s", err, w.Body.String())
	}
	return out
}

func itoa(n int) string {
	raw, _ := json.Marshal(n)
	return string(raw)
}
