package wg

import "testing"

func TestParseTransferOutput(t *testing.T) {
	out := parseTransferOutput("" +
		"pubkeyA 100 200\n" +
		"pubkeyB 0 0\n" +
		"badline\n" +
		"pubkeyC notanumber 1\n" +
		"pubkeyD 50 75 extra\n")
	if len(out) != 3 {
		t.Fatalf("len = %d, want 3: %+v", len(out), out)
	}
	if out["pubkeyA"] != (Transfer{Rx: 100, Tx: 200}) {
		t.Fatalf("A = %+v", out["pubkeyA"])
	}
	if out["pubkeyB"] != (Transfer{Rx: 0, Tx: 0}) {
		t.Fatalf("B = %+v", out["pubkeyB"])
	}
	if out["pubkeyD"] != (Transfer{Rx: 50, Tx: 75}) {
		t.Fatalf("D = %+v", out["pubkeyD"])
	}
	if _, ok := out["pubkeyC"]; ok {
		t.Fatal("malformed line should be skipped")
	}
}
