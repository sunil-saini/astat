package cmd

import (
	"fmt"

	"github.com/sunil-saini/astat/internal/refresh"
)

// plainTracker implements refresh.Tracker for non-interactive terminals
type plainTracker struct {
	prefix string
}

func newPlainTracker(prefix string) refresh.Tracker {
	return &plainTracker{prefix: prefix}
}

func (p *plainTracker) Update(msg string) {
	// Don't print intermediate updates in non-interactive mode
	_ = msg
}

func (p *plainTracker) Success(msg string) {
	_ = msg
	fmt.Printf("✓ %s\n", p.prefix)
}

func (p *plainTracker) Error(msg string) {
	fmt.Printf("✗ (%s)\n", msg)
}
