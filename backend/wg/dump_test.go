package wg

import "testing"

func TestParseDump(t *testing.T) {
	out := "privkey\tpubkey\t51820\toff\n" +
		"peerA\t(none)\t203.0.113.7:40001\t10.0.0.2/32\t1700000000\t100\t200\t25\n" +
		"peerB\tpsk\t(none)\t10.0.0.3/32\t0\t0\t0\toff\n" +
		"badline\n" +
		"peerC\t(none)\t(none)\t10.0.0.4/32\tnope\t1\t2\toff\n" +
		"peerD\t(none)\t[2001:db8::1]:51820\t10.0.0.5/32\t1700000100\t50\t75\toff\n"

	got := parseDump(out)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3: %+v", len(got), got)
	}

	a := got["peerA"]
	if a.Endpoint != "203.0.113.7:40001" || a.LatestHandshake != 1700000000 || a.Transfer != (Transfer{Rx: 100, Tx: 200}) {
		t.Fatalf("A = %+v", a)
	}
	b := got["peerB"]
	if b.Endpoint != "" || b.LatestHandshake != 0 || b.Transfer != (Transfer{}) {
		t.Fatalf("B = %+v", b)
	}
	if d := got["peerD"]; d.Endpoint != "[2001:db8::1]:51820" || d.Transfer != (Transfer{Rx: 50, Tx: 75}) {
		t.Fatalf("D = %+v", d)
	}
	if _, ok := got["peerC"]; ok {
		t.Fatal("malformed line should be skipped")
	}
}

func TestParseDumpInterfaceOnly(t *testing.T) {
	if got := parseDump("privkey\tpubkey\t51820\toff\n"); len(got) != 0 {
		t.Fatalf("interface line only should yield no peers, got %+v", got)
	}
	if got := parseDump(""); len(got) != 0 {
		t.Fatalf("empty output should yield no peers, got %+v", got)
	}
}

func TestTransfers(t *testing.T) {
	got := Transfers(map[string]PeerState{
		"a": {Endpoint: "1.2.3.4:5", Transfer: Transfer{Rx: 1, Tx: 2}},
	})
	if got["a"] != (Transfer{Rx: 1, Tx: 2}) {
		t.Fatalf("got %+v", got)
	}
}

func TestHandshakeAge(t *testing.T) {
	if HandshakeAge(0) != -1 {
		t.Fatal("never handshaked should be -1")
	}
	if age := HandshakeAge(1); age < 1 {
		t.Fatalf("old handshake age = %d", age)
	}
}
