package sim

import (
	"sync"
	"time"
)

// Clock abstracts time so the engine can be driven deterministically in
// tests instead of depending on real sleeps.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// RealClock is the production Clock, backed by time.Now.
var RealClock Clock = realClock{}

// ManualClock is a Clock a test can advance explicitly.
type ManualClock struct {
	mu  sync.Mutex
	now time.Time
}

// NewManualClock returns a ManualClock starting at start.
func NewManualClock(start time.Time) *ManualClock {
	return &ManualClock{now: start}
}

// Now returns the clock's current time.
func (c *ManualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock's current time forward by d.
func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
