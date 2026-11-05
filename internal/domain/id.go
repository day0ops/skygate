package domain

import (
	"crypto/rand"
	"fmt"
)

// NewID returns a short random identifier prefixed with p, e.g. "flt_3f9a2c1b".
func NewID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%x", prefix, b)
}
