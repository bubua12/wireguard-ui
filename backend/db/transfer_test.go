package db

import (
	"path/filepath"
	"testing"
	"time"
	"wireguard-ui/model"
	"wireguard-ui/wg"
)

func TestTransferDelta(t *testing.T) {
	if g := transferDelta(0, 100, false); g != 100 {
		t.Fatalf("first seen = %d", g)
	}
	if g := transferDelta(100, 150, true); g != 50 {
		t.Fatalf("increase = %d", g)
	}
	if g := transferDelta(100, 100, true); g != 0 {
		t.Fatalf("unchanged = %d", g)
	}
	if g := transferDelta(100, 10, true); g != 10 {
		t.Fatalf("reset = %d", g)
	}
}

func TestMergePending(t *testing.T) {
	days := []TrafficDay{{Day: "2026-08-19", Rx: 5, Tx: 6}}
	got := mergePending(days, 1, 2, "2026-08-20")
	if len(got) != 2 || got[0].Day != "2026-08-20" || got[0].Rx != 1 || got[0].Tx != 2 {
		t.Fatalf("insert today = %+v", got)
	}

	days = []TrafficDay{{Day: "2026-08-20", Rx: 5, Tx: 6}, {Day: "2026-08-19", Rx: 1, Tx: 1}}
	got = mergePending(days, 3, 4, "2026-08-20")
	if got[0].Rx != 8 || got[0].Tx != 10 {
		t.Fatalf("add to today = %+v", got[0])
	}

	unchanged := []TrafficDay{{Day: "2026-08-20", Rx: 5, Tx: 6}}
	got = mergePending(unchanged, 0, 0, "2026-08-20")
	if got[0].Rx != 5 || len(got) != 1 {
		t.Fatalf("no pending should keep original, got %+v", got)
	}
}

func setupTransferDB(t *testing.T) *model.Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "t.db")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	server := &model.Server{
		Name: "test", Interface: "wg0", PrivateKey: "priv", PublicKey: "pub",
		Address: "10.0.0.1/24", ListenPort: 51820, Endpoint: "vpn.example.com:51820",
		DNS: "8.8.8.8", MTU: 1420,
	}
	if err := CreateServer(server); err != nil {
		t.Fatal(err)
	}
	p := &model.Peer{ServerID: server.ID, Name: "a", PrivateKey: "k1", PublicKey: "peerA", Enabled: true}
	if err := AllocateAndCreatePeer(server, p, "10.0.0.2/32"); err != nil {
		t.Fatal(err)
	}
	return server
}

func TestUsageBeforeFirstSample(t *testing.T) {
	setupTransferDB(t)
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.Local)
	live := map[string]wg.Transfer{"peerA": {Rx: 42, Tx: 7}}

	usage, err := UsageForKeys([]string{"peerA"}, live)
	if err != nil {
		t.Fatal(err)
	}
	if usage["peerA"] != (wg.Transfer{Rx: 42, Tx: 7}) {
		t.Fatalf("usage before sample = %+v", usage["peerA"])
	}

	report, err := GetPeerTrafficReport("peerA", live["peerA"], now)
	if err != nil {
		t.Fatal(err)
	}
	if report.Rx != 42 || report.Tx != 7 || len(report.Days) != 1 || report.Days[0].Day != "2026-08-20" {
		t.Fatalf("report before sample = %+v", report)
	}
}

func TestApplyTransferSampleAndUsage(t *testing.T) {
	setupTransferDB(t)
	day1 := time.Date(2026, 8, 19, 12, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 8, 20, 9, 0, 0, 0, time.Local)

	live := map[string]wg.Transfer{"peerA": {Rx: 1000, Tx: 200}}
	if err := ApplyTransferSample(live, day1); err != nil {
		t.Fatal(err)
	}

	usage, err := UsageForKeys([]string{"peerA"}, live)
	if err != nil {
		t.Fatal(err)
	}
	if usage["peerA"] != (wg.Transfer{Rx: 1000, Tx: 200}) {
		t.Fatalf("after first sample = %+v", usage["peerA"])
	}

	live["peerA"] = wg.Transfer{Rx: 1500, Tx: 250}
	if err := ApplyTransferSample(live, day2); err != nil {
		t.Fatal(err)
	}

	report, err := GetPeerTrafficReport("peerA", live["peerA"], day2)
	if err != nil {
		t.Fatal(err)
	}
	if report.Rx != 1500 || report.Tx != 250 {
		t.Fatalf("report totals = %+v", report)
	}
	if len(report.Days) != 2 {
		t.Fatalf("days = %+v", report.Days)
	}
	if report.Days[0].Day != "2026-08-20" || report.Days[0].Rx != 500 || report.Days[0].Tx != 50 {
		t.Fatalf("day2 = %+v", report.Days[0])
	}
	if report.Days[1].Day != "2026-08-19" || report.Days[1].Rx != 1000 || report.Days[1].Tx != 200 {
		t.Fatalf("day1 = %+v", report.Days[1])
	}

	// Counter reset (interface restart): current values are a new session.
	live["peerA"] = wg.Transfer{Rx: 40, Tx: 10}
	if err := ApplyTransferSample(live, day2); err != nil {
		t.Fatal(err)
	}
	report, err = GetPeerTrafficReport("peerA", live["peerA"], day2)
	if err != nil {
		t.Fatal(err)
	}
	if report.Rx != 1540 || report.Tx != 260 {
		t.Fatalf("after reset totals = %+v", report)
	}

	// Unsampled live delta is included in usage.
	live["peerA"] = wg.Transfer{Rx: 60, Tx: 15}
	usage, err = UsageForKeys([]string{"peerA"}, live)
	if err != nil {
		t.Fatal(err)
	}
	if usage["peerA"] != (wg.Transfer{Rx: 1560, Tx: 265}) {
		t.Fatalf("pending usage = %+v", usage["peerA"])
	}
}

func TestApplyTransferSamplePrunesMissingPeers(t *testing.T) {
	setupTransferDB(t)
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.Local)
	live := map[string]wg.Transfer{
		"peerA":     {Rx: 10, Tx: 1},
		"ghostPeer": {Rx: 99, Tx: 9},
	}
	if err := ApplyTransferSample(live, now); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM peer_transfer_cursor WHERE public_key = 'ghostPeer'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("ghost cursor rows = %d", n)
	}
}
