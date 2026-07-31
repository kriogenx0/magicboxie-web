package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"magicbox/internal/auth"
)

var mediaBrowserTokenRe = regexp.MustCompile(`Token="([^"]*)"`)

// RequireAuth accepts three token transports, in priority order:
//  1. A Jellyfin-style "Authorization: MediaBrowser ...Token="..."" header
//     (what magicbox-appletv's JellyfinClient sends).
//  2. A plain "Authorization: Bearer <token>" header (simpler clients).
//  3. A "?api_key=<token>" query param -- Jellyfin's convention for URLs
//     that can't carry headers (<video>/<img> tags, AVPlayer, EventSource).
func RequireAuth(manager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		if err := manager.VerifyToken(token); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "MediaBrowser ") {
		if m := mediaBrowserTokenRe.FindStringSubmatch(header); m != nil {
			return m[1]
		}
	}
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return c.Query("api_key")
}
