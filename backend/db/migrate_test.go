package db

import (
	"path/filepath"
	"testing"
	"time"
)

// A legacy row with an unusable interface name used to hang startup: the repair
// ran while the cursor still held the pool's only connection.
func TestInitRepairsInvalidInterfaceName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	if _, err := DB.Exec(`INSERT INTO servers (name, interface, private_key, public_key, address, endpoint) VALUES ('x', 'bubua12.com', 'k', 'p', '10.0.0.1/24', 'vpn.example.com:51820')`); err != nil {
		t.Fatal(err)
	}
	if err := Close(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- Init(path) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Init deadlocked while repairing an invalid interface name")
	}

	s, err := GetFirstServer()
	if err != nil {
		t.Fatal(err)
	}
	if s.Interface != "wg0" {
		t.Fatalf("interface = %q, want wg0", s.Interface)
	}
}
