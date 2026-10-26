package db

import (
	"testing"
	"time"
	"wireguard-ui/wg"
)

func TestGetTrafficHistory(t *testing.T) {
	setupTransferDB(t) // one peer: peerA
	day1 := time.Date(2026, 8, 18, 12, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 8, 20, 9, 0, 0, 0, time.Local)

	if err := ApplyTransferSample(map[string]wg.Transfer{"peerA": {Rx: 1000, Tx: 200}}, day1); err != nil {
		t.Fatal(err)
	}
	live := map[string]wg.Transfer{"peerA": {Rx: 1500, Tx: 260}}
	if err := ApplyTransferSample(map[string]wg.Transfer{"peerA": {Rx: 1300, Tx: 250}}, day2); err != nil {
		t.Fatal(err)
	}

	got, err := GetTrafficHistory([]string{"peerA"}, live, 4, day2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("len = %d: %+v", len(got), got)
	}
	want := []TrafficDay{
		{Day: "2026-08-17"},
		{Day: "2026-08-18", Rx: 1000, Tx: 200},
		{Day: "2026-08-19"},
		// 300/50 sampled today plus 200/10 still pending in the live counters.
		{Day: "2026-08-20", Rx: 500, Tx: 60},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("day %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	// Rows of peers that are no longer listed do not leak into the totals.
	got, err = GetTrafficHistory([]string{"someone-else"}, nil, 4, day2)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range got {
		if d.Rx != 0 || d.Tx != 0 {
			t.Fatalf("unexpected traffic for unlisted peer: %+v", d)
		}
	}

	// Window is clamped to something sane.
	got, err = GetTrafficHistory(nil, nil, 100000, day2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != transferRetentionDays {
		t.Fatalf("clamped len = %d", len(got))
	}
}
