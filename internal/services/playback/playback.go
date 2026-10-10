// Package playback tracks whether anyone is watching or listening, so
// background work (480p/1080p transcodes) can stand down and leave the
// Pi's CPU to the stream.
package playback

import (
	"sync"
	"time"
)

// Grace is how long after a stream request ends playback still counts as
// happening. Players fetch a file as many short Range requests with idle
// gaps while their buffer is full, so an open request alone can't tell
// whether someone is still watching.
const Grace = 60 * time.Second

type Tracker struct {
	mu     sync.Mutex
	active int
	last   time.Time
}

func NewTracker() *Tracker { return &Tracker{} }

// Begin marks a stream request as started; call the returned func when it ends.
func (t *Tracker) Begin() (end func()) {
	t.mu.Lock()
	t.active++
	t.mu.Unlock()
	return func() {
		t.mu.Lock()
		t.active--
		t.last = time.Now()
		t.mu.Unlock()
	}
}

// Playing reports whether a stream is open now or one ended within Grace.
func (t *Tracker) Playing() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active > 0 || (!t.last.IsZero() && time.Since(t.last) < Grace)
}
