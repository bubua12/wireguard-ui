package wg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateInterfaceName(t *testing.T) {
	ok := []string{"wg0", "wg-office", "tun0"}
	for _, n := range ok {
		if err := ValidateInterfaceName(n); err != nil {
			t.Fatalf("%s should be valid: %v", n, err)
		}
	}
	bad := []string{"", "My VPN Server", "../etc", "1wg", "thisnameiswaytoolong", "bubua12.com"}
	for _, n := range bad {
		if err := ValidateInterfaceName(n); err == nil {
			t.Fatalf("%q should be invalid", n)
		}
	}
}

func TestSanitizeDownloadName(t *testing.T) {
	if got := SanitizeDownloadName(`foo"bar.conf`); got == `foo"bar.conf` {
		t.Fatal("quotes should be stripped")
	}
	if got := SanitizeDownloadName("iPhone 15"); got != "iPhone_15" {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeDownloadName(""); got != "client" {
		t.Fatalf("empty = %q", got)
	}
}

func TestValidateImportPath(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "wg0.conf")
	if err := os.WriteFile(good, []byte("[Interface]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ValidateImportPath(dir, good)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != good {
		t.Fatalf("got %s", got)
	}

	if _, err := ValidateImportPath(dir, filepath.Join(dir, "..", "passwd")); err == nil {
		t.Fatal("path escape should fail")
	}
	if _, err := ValidateImportPath(dir, "/etc/passwd"); err == nil {
		t.Fatal("outside dir should fail")
	}
}

func TestValidateEndpointAndAddress(t *testing.T) {
	if err := ValidateEndpoint("vpn.example.com:51820"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEndpoint("vpn.example.com"); err == nil {
		t.Fatal("endpoint without port should fail")
	}
	if err := ValidateServerAddress("10.0.0.1/24"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateServerAddress("10.0.0.1"); err == nil {
		t.Fatal("bare ip should fail")
	}
	if err := ValidateListenPort(51820); err != nil {
		t.Fatal(err)
	}
	if err := ValidateListenPort(0); err == nil {
		t.Fatal("port 0 should fail")
	}
}
