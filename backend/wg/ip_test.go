package wg

import "testing"

func TestNextAvailableIP(t *testing.T) {
	got, err := NextAvailableIP("10.0.0.1/24", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.0.0.2/32" {
		t.Fatalf("first ip = %q, want 10.0.0.2/32", got)
	}

	got, err = NextAvailableIP("10.0.0.1/24", []string{"10.0.0.2/32", "10.0.0.3/32"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.0.0.4/32" {
		t.Fatalf("next ip = %q, want 10.0.0.4/32", got)
	}

	// holes should be reused
	got, err = NextAvailableIP("10.0.0.1/24", []string{"10.0.0.3/32", "10.0.0.4/32"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.0.0.2/32" {
		t.Fatalf("hole fill = %q, want 10.0.0.2/32", got)
	}

	_, err = NextAvailableIP("10.0.0.1/30", []string{"10.0.0.2/32"})
	if err == nil {
		t.Fatal("expected no-available-ip error for exhausted /30")
	}

	_, err = NextAvailableIP("not-a-cidr", nil)
	if err == nil {
		t.Fatal("expected error for invalid cidr")
	}
}

func TestValidateClientIP(t *testing.T) {
	if err := ValidateClientIP("10.0.0.1/24", "10.0.0.2/32"); err != nil {
		t.Fatalf("valid ip rejected: %v", err)
	}
	cases := []string{
		"10.0.0.2/24",
		"10.0.0.1/32",
		"10.0.0.0/32",
		"10.0.0.255/32",
		"10.1.0.2/32",
		"abc/32",
		"10.0.0.2",
	}
	for _, c := range cases {
		if err := ValidateClientIP("10.0.0.1/24", c); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}

func TestCanonicalClientIP(t *testing.T) {
	got, err := CanonicalClientIP("10.0.0.2/32")
	if err != nil || got != "10.0.0.2/32" {
		t.Fatalf("got %q %v", got, err)
	}
}
