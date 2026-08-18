package wg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wg0.conf")
	content := `[Interface]
PrivateKey = serverpriv
Address = 10.8.0.1/24
ListenPort = 51820
DNS = 1.1.1.1
MTU = 1420

[Peer]
PublicKey = peer1pub
PresharedKey = psk1
AllowedIPs = 10.8.0.2/32
PersistentKeepalive = 25

[Peer]
PublicKey = peer2pub
AllowedIPs = 10.8.0.3/32
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseConfigFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Interface.Address != "10.8.0.1/24" || parsed.Interface.ListenPort != 51820 {
		t.Fatalf("interface parse mismatch: %+v", parsed.Interface)
	}
	if len(parsed.Peers) != 2 {
		t.Fatalf("peers = %d", len(parsed.Peers))
	}
	if parsed.Peers[0].PresharedKey != "psk1" || parsed.Peers[1].PublicKey != "peer2pub" {
		t.Fatalf("peer fields: %+v", parsed.Peers)
	}
}
