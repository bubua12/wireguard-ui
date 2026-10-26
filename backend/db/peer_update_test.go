package db

import (
	"testing"
	"wireguard-ui/model"
)

func TestUpdatePeer(t *testing.T) {
	server := setupTransferDB(t) // peerA already holds 10.0.0.2/32

	b := &model.Peer{ServerID: server.ID, Name: "b", PrivateKey: "k2", PublicKey: "peerB", Enabled: true, PersistentKeepalive: 25}
	if err := AllocateAndCreatePeer(server, b, ""); err != nil {
		t.Fatal(err)
	}
	if b.AllowedIPs != "10.0.0.3/32" {
		t.Fatalf("b ip = %s", b.AllowedIPs)
	}

	// Plain field edit keeps the address.
	b.Name, b.PersistentKeepalive = "renamed", 0
	if err := UpdatePeer(b, ""); err != nil {
		t.Fatal(err)
	}
	got, err := GetPeer(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "renamed" || got.PersistentKeepalive != 0 || got.AllowedIPs != "10.0.0.3/32" {
		t.Fatalf("after rename = %+v", got)
	}

	// Moving onto another peer's address is refused and leaves the row alone.
	if err := UpdatePeer(got, "10.0.0.2/32"); err == nil {
		t.Fatal("expected duplicate ip error")
	}
	if again, _ := GetPeer(b.ID); again.AllowedIPs != "10.0.0.3/32" {
		t.Fatalf("ip changed despite error: %s", again.AllowedIPs)
	}

	// Moving to a free address works, and the old one becomes reusable.
	got, _ = GetPeer(b.ID)
	if err := UpdatePeer(got, "10.0.0.50/32"); err != nil {
		t.Fatal(err)
	}
	if again, _ := GetPeer(b.ID); again.AllowedIPs != "10.0.0.50/32" {
		t.Fatalf("ip = %s", again.AllowedIPs)
	}
	c := &model.Peer{ServerID: server.ID, Name: "c", PrivateKey: "k3", PublicKey: "peerC", Enabled: true}
	if err := AllocateAndCreatePeer(server, c, "10.0.0.3/32"); err != nil {
		t.Fatalf("freed address should be reusable: %v", err)
	}

	// Keeping the same address is not a conflict with itself.
	cur, _ := GetPeer(c.ID)
	if err := UpdatePeer(cur, "10.0.0.3/32"); err != nil {
		t.Fatalf("same ip: %v", err)
	}
}

// Peers come back in creation order, which keeps the generated server config
// and the default client list stable between requests.
func TestGetPeersByServerOrder(t *testing.T) {
	server := setupTransferDB(t)
	for _, key := range []string{"zzz", "mmm", "aaa"} {
		p := &model.Peer{ServerID: server.ID, Name: key, PrivateKey: "k" + key, PublicKey: key, Enabled: true}
		if err := AllocateAndCreatePeer(server, p, ""); err != nil {
			t.Fatal(err)
		}
	}
	peers, err := GetPeersByServer(server.ID)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range peers {
		got = append(got, p.PublicKey)
	}
	want := []string{"peerA", "zzz", "mmm", "aaa"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}
