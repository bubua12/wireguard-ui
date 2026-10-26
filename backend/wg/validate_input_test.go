package wg

import (
	"strings"
	"testing"
	"time"
	"wireguard-ui/model"
)

func TestValidateEndpoint(t *testing.T) {
	ok := []string{"vpn.example.com:51820", "203.0.113.7:51820", "[2001:db8::1]:51820", "my_host:1"}
	for _, e := range ok {
		if err := ValidateEndpoint(e); err != nil {
			t.Fatalf("%q should be valid: %v", e, err)
		}
	}
	bad := []string{
		"", "vpn.example.com", ":51820", "vpn.example.com:0", "vpn.example.com:70000",
		"vpn.example.com:abc", "bad host:51820", "evil\nPostUp = id:51820", "a/b:51820",
	}
	for _, e := range bad {
		if err := ValidateEndpoint(e); err == nil {
			t.Fatalf("%q should be invalid", e)
		}
	}
}

func TestValidateDNS(t *testing.T) {
	ok := []string{"", "  ", "8.8.8.8", "1.1.1.1, 8.8.8.8", "2606:4700:4700::1111", "dns.example.com", "10.0.0.1,example.lan"}
	for _, d := range ok {
		if err := ValidateDNS(d); err != nil {
			t.Fatalf("%q should be valid: %v", d, err)
		}
	}
	bad := []string{"1.1.1.1,", ",8.8.8.8", "8.8.8.8\nPostUp = id", "bad dns", "8.8.8.8;rm", "-bad.example.com"}
	for _, d := range bad {
		if err := ValidateDNS(d); err == nil {
			t.Fatalf("%q should be invalid", d)
		}
	}
	if got := NormalizeDNS(" 1.1.1.1 ,8.8.8.8"); got != "1.1.1.1, 8.8.8.8" {
		t.Fatalf("normalized = %q", got)
	}
}

func TestValidateMTU(t *testing.T) {
	for _, m := range []int{0, 576, 1420, 9000} {
		if err := ValidateMTU(m); err != nil {
			t.Fatalf("%d should be valid: %v", m, err)
		}
	}
	for _, m := range []int{-1, 100, 575, 9001} {
		if err := ValidateMTU(m); err == nil {
			t.Fatalf("%d should be invalid", m)
		}
	}
}

func TestValidateKey(t *testing.T) {
	good := "YNqHbfBQKaGvzefSZ6ybrwHsBqUPyOcACNZZ0b0+0nQ="
	if err := ValidateKey(good); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
	for _, k := range []string{"", "short", good[:43], "--help", strings.Repeat("A", 44)} {
		if err := ValidateKey(k); err == nil {
			t.Fatalf("%q should be invalid", k)
		}
	}
}

func TestGenerateClientConfigOptionalLines(t *testing.T) {
	server := &model.Server{PublicKey: "spub", Address: "10.0.0.1/24", Endpoint: "vpn.example.com:51820"}
	peer := &model.Peer{PrivateKey: "cpriv", AllowedIPs: "10.0.0.2/32"}

	cfg := GenerateClientConfig(server, peer)
	for _, unwanted := range []string{"DNS =", "MTU =", "PersistentKeepalive", "PresharedKey"} {
		if strings.Contains(cfg, unwanted) {
			t.Fatalf("%q should be omitted when unset:\n%s", unwanted, cfg)
		}
	}

	server.DNS, server.MTU = "1.1.1.1", 1380
	peer.PersistentKeepalive, peer.PresharedKey = 25, "psk"
	cfg = GenerateClientConfig(server, peer)
	for _, want := range []string{"DNS = 1.1.1.1", "MTU = 1380", "PersistentKeepalive = 25", "PresharedKey = psk"} {
		if !strings.Contains(cfg, want) {
			t.Fatalf("missing %q:\n%s", want, cfg)
		}
	}
}

func TestRunCmdTimeout(t *testing.T) {
	start := time.Now()
	_, err := runCmd(150*time.Millisecond, "", "sleep", "5")
	if err == nil || !strings.Contains(err.Error(), "超时") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if time.Since(start) > 4*time.Second {
		t.Fatalf("command was not killed promptly: %v", time.Since(start))
	}
}

func TestRunCmdMissingBinary(t *testing.T) {
	_, err := runCmd(time.Second, "", "definitely-not-a-real-binary-xyz")
	if err == nil || !strings.Contains(err.Error(), "未找到") {
		t.Fatalf("expected friendly not-found error, got %v", err)
	}
}
