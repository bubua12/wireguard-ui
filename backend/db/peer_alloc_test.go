package db

import (
	"path/filepath"
	"testing"
	"wireguard-ui/model"
)

func TestAllocateAndCreatePeer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.db")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}

	server := &model.Server{
		Name:       "test",
		Interface:  "wg0",
		PrivateKey: "priv",
		PublicKey:  "pub",
		Address:    "10.0.0.1/24",
		ListenPort: 51820,
		Endpoint:   "vpn.example.com:51820",
		DNS:        "8.8.8.8",
		MTU:        1420,
	}
	if err := CreateServer(server); err != nil {
		t.Fatal(err)
	}

	p1 := &model.Peer{ServerID: server.ID, Name: "a", PrivateKey: "k1", PublicKey: "p1", Enabled: true}
	if err := AllocateAndCreatePeer(server, p1, ""); err != nil {
		t.Fatal(err)
	}
	if p1.AllowedIPs != "10.0.0.2/32" {
		t.Fatalf("auto ip = %s", p1.AllowedIPs)
	}

	p2 := &model.Peer{ServerID: server.ID, Name: "b", PrivateKey: "k2", PublicKey: "p2", Enabled: true}
	if err := AllocateAndCreatePeer(server, p2, "10.0.0.10/32"); err != nil {
		t.Fatal(err)
	}
	if p2.AllowedIPs != "10.0.0.10/32" {
		t.Fatalf("custom ip = %s", p2.AllowedIPs)
	}

	p3 := &model.Peer{ServerID: server.ID, Name: "c", PrivateKey: "k3", PublicKey: "p3", Enabled: true}
	if err := AllocateAndCreatePeer(server, p3, "10.0.0.10/32"); err == nil {
		t.Fatal("expected duplicate ip error")
	}

	p4 := &model.Peer{ServerID: server.ID, Name: "d", PrivateKey: "k4", PublicKey: "p4", Enabled: true}
	if err := AllocateAndCreatePeer(server, p4, "10.0.0.1/32"); err == nil {
		t.Fatal("expected server address rejection")
	}

	p5 := &model.Peer{ServerID: server.ID, Name: "e", PrivateKey: "k5", PublicKey: "p5", Enabled: true}
	if err := AllocateAndCreatePeer(server, p5, ""); err != nil {
		t.Fatal(err)
	}
	if p5.AllowedIPs != "10.0.0.3/32" {
		t.Fatalf("next hole-aware ip = %s, want 10.0.0.3/32", p5.AllowedIPs)
	}
}
