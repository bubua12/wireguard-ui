package wg

import (
	"strings"
	"testing"
	"wireguard-ui/model"
)

func TestGenerateServerConfigNAT(t *testing.T) {
	server := &model.Server{
		PrivateKey: "skey",
		Address:    "10.0.0.1/24",
		ListenPort: 51820,
		EnableNAT:  true,
		MTU:        1420,
	}
	peers := []model.Peer{{
		PublicKey:  "pkey",
		AllowedIPs: "10.0.0.2/32",
		Enabled:    true,
	}}
	cfg := GenerateServerConfig(server, peers)
	if !strings.Contains(cfg, "PostUp =") || !strings.Contains(cfg, "-s 10.0.0.0/24") {
		t.Fatalf("missing NAT rules:\n%s", cfg)
	}
	if strings.Contains(cfg, "%%i") {
		t.Fatal("%%i was not interpolated")
	}
	if !strings.Contains(cfg, "-i %i") {
		t.Fatalf("expected %%i interface token:\n%s", cfg)
	}
}

func TestGenerateClientConfigTunnelModes(t *testing.T) {
	server := &model.Server{
		PublicKey: "spub",
		Address:   "10.0.0.1/24",
		Endpoint:  "vpn.example.com:51820",
		DNS:       "8.8.8.8",
	}
	peer := &model.Peer{PrivateKey: "cpriv", AllowedIPs: "10.0.0.2/32", PersistentKeepalive: 25}

	split := GenerateClientConfig(server, peer)
	if !strings.Contains(split, "AllowedIPs = 10.0.0.0/24") {
		t.Fatalf("split tunnel:\n%s", split)
	}

	server.FullTunnel = true
	full := GenerateClientConfig(server, peer)
	if !strings.Contains(full, "AllowedIPs = 0.0.0.0/0, ::/0") {
		t.Fatalf("full tunnel:\n%s", full)
	}
}
