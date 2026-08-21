package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"magicboxie/internal/services/events"
)

// Events streams job_progress / job_completed / job_failed updates over
// Server-Sent Events. The frontend still does a normal REST fetch for full
// state on load; this stream only layers live deltas on top, so a dropped
// connection (browser auto-reconnects EventSource) is self-healing.
func Events(hub *events.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		ch := hub.Subscribe()
		defer hub.Unsubscribe(ch)

		c.Writer.WriteHeader(http.StatusOK)
		flusher.Flush()

		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(evt)
				if err != nil {
					continue
				}
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				flusher.Flush()
			case <-c.Request.Context().Done():
				return
			}
		}
	}
}
