package middleware

import (
	"github.com/gin-gonic/gin"

	"magicboxie/internal/services/playback"
)

// TrackPlayback records each request as playback activity for the duration
// of the response. A nil tracker makes it a no-op.
func TrackPlayback(t *playback.Tracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if t != nil {
			defer t.Begin()()
		}
		c.Next()
	}
}
