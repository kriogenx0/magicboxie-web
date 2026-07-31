package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PlayingProgress and PlayingStopped are Jellyfin's playback-reporting
// endpoints. magicbox-appletv's client currently sends no body on these
// (confirmed during API investigation -- JellyfinRouter.httpBody returns
// nil for both cases), so these handlers acknowledge whatever they get
// rather than requiring specific fields. Richer position tracking can be
// layered on later without a contract change.
func PlayingProgress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func PlayingStopped(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
