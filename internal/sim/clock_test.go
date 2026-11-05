package sim

import (
	"testing"
	"time"
)

func TestManualClockAdvance(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := NewManualClock(start)

	if !c.Now().Equal(start) {
		t.Fatalf("got %v, want %v", c.Now(), start)
	}

	c.Advance(5 * time.Second)
	want := start.Add(5 * time.Second)
	if !c.Now().Equal(want) {
		t.Fatalf("got %v, want %v", c.Now(), want)
	}
}
