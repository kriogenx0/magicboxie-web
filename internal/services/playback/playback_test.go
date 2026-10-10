package playback

import "testing"

func TestTrackerPlayingWhileActiveAndDuringGrace(t *testing.T) {
	tr := NewTracker()
	if tr.Playing() {
		t.Fatal("idle tracker reports playing")
	}
	end := tr.Begin()
	if !tr.Playing() {
		t.Fatal("open stream not counted")
	}
	end()
	if !tr.Playing() {
		t.Fatal("grace period after a stream ends not counted")
	}
}
