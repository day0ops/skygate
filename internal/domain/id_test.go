package domain

import "testing"

func TestNewIDIsUniqueAndPrefixed(t *testing.T) {
	a := NewID("flt")
	b := NewID("flt")

	if a == b {
		t.Fatalf("expected distinct IDs, got %q twice", a)
	}
	if len(a) <= len("flt_") {
		t.Fatalf("expected a suffix after the prefix, got %q", a)
	}
	if a[:4] != "flt_" {
		t.Fatalf("expected prefix \"flt_\", got %q", a)
	}
}
