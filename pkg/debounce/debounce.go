package debounce

import (
	"sync"
	"time"
)

// New creates a debounced function that takes another functions as its argument.
// This function will be called when the debounced function stops being called
// for the given duration.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func New(after time.Duration) *Limiter {
	return &Limiter{after: after}
}

type Limiter struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

func (d *Limiter) Call(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}
