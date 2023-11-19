package debounce

import (
	"sync"
	"time"
)

// New creates a debounced function that takes another functions as its argument.
func New(after time.Duration) *Limiter {
	return &Limiter{after: after}
}

type Limiter struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

// Fn calls the given function if the debounced function is not called
// for the given duration. If the debounced function is called before the
// duration expires, the given function will not be called and the timer
// will be reset.
// This function will be called when the debounced function stops being called
// for the given duration.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func (d *Limiter) Fn(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}

// Call schedules the given function to be called after duration.
// If Call is called again before the duration expires, it will be a No-Op.
func (d *Limiter) Call(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer == nil {
		d.timer = time.AfterFunc(d.after, func() {
			f()
			d.mu.Lock()
			defer d.mu.Unlock()
			d.timer = nil
		})
	}
}
