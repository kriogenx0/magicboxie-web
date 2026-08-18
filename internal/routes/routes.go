package routes

import (
	"github.com/gin-gonic/gin"

	"magicbox/internal/auth"
	"magicbox/internal/controllers"
	"magicbox/internal/middleware"
	"magicbox/internal/services/events"
)

type Dependencies struct {
	AuthManager       *auth.Manager
	AuthController    *controllers.AuthController
	ItemsController   *controllers.ItemsController
	VideosController  *controllers.VideosController
	AudioController   *controllers.AudioController
	UploadsController *controllers.UploadsController
	EventsHub         *events.Hub
}

// Register wires two surfaces onto the router:
//
//  1. A Jellyfin-compatible surface at bare paths (/System, /Users, /Items,
//     /Videos, /Sessions) matching real Jellyfin's conventions exactly, so
//     magicbox-appletv (already a partial Jellyfin client) and any generic
//     Jellyfin client work against this server unmodified.
//  2. MagicBox-specific extensions with no Jellyfin equivalent (chunked
//     upload, library scan trigger, SSE job-progress, manual TMDB
//     re-match), namespaced under /api/* so they're clearly separate
//     from the standard surface.
func Register(router *gin.Engine, deps Dependencies) {
	router.GET("/System/Info/Public", deps.AuthController.SystemInfoPublic)
	router.POST("/Users/AuthenticateByName", deps.AuthController.AuthenticateByName)

	// Unauthenticated check-in for magicboxie-device Pis (see
	// player_app/views/home_sync_service.py in that repo) -- a lower-friction
	// alternative to the Jellyfin login flow above for a headless device that
	// boots straight into opportunistic sync.
	router.POST("/devices/register", deps.ItemsController.RegisterDevice)

	// Images are unauthenticated, matching real Jellyfin's convention (so
	// <img> tags never need a token).
	router.GET("/Items/:itemId/Images/Primary", deps.ItemsController.PrimaryImage)
	router.GET("/Items/:itemId/Images/Backdrop/:index", deps.ItemsController.BackdropImage)
	router.GET("/Items/:itemId/Images/Thumbnail/:index", deps.ItemsController.ThumbnailCandidateImage)

	router.GET("/api/health", controllers.Health)

	authorized := router.Group("")
	authorized.Use(middleware.RequireAuth(deps.AuthManager))
	{
		authorized.GET("/Users/:userId/Views", deps.ItemsController.Views)
		authorized.GET("/Users/:userId/Items", deps.ItemsController.List)
		authorized.GET("/Users/:userId/Items/Latest", deps.ItemsController.Latest)
		authorized.GET("/Users/:userId/Items/:itemId", deps.ItemsController.Detail)

		authorized.POST("/Items/:itemId/PlaybackInfo", deps.ItemsController.PlaybackInfo)

		authorized.GET("/Videos/:itemId/stream", deps.VideosController.Stream)
		authorized.HEAD("/Videos/:itemId/stream", deps.VideosController.Stream)

		authorized.GET("/Audio/:itemId/stream", deps.AudioController.Stream)
		authorized.HEAD("/Audio/:itemId/stream", deps.AudioController.Stream)

		authorized.POST("/Sessions/Playing/Progress", controllers.PlayingProgress)
		authorized.POST("/Sessions/Playing/Stopped", controllers.PlayingStopped)

		api := authorized.Group("/api")
		{
			api.POST("/library/scan", deps.ItemsController.Scan)
			api.POST("/library/music/scan", deps.ItemsController.MusicScan)
			api.POST("/items/:itemId/match", deps.ItemsController.Match)
			api.GET("/items/:itemId/thumbnails", deps.ItemsController.ThumbnailCandidates)
			api.POST("/items/:itemId/thumbnails/select", deps.ItemsController.SelectThumbnail)
			api.POST("/items/:itemId/sync", deps.ItemsController.SetDeviceSync)
			api.GET("/items/search", deps.ItemsController.Search)

			api.POST("/uploads", deps.UploadsController.Create)
			api.GET("/uploads/:id", deps.UploadsController.Status)
			api.PUT("/uploads/:id/chunk", deps.UploadsController.Chunk)
			api.POST("/uploads/:id/complete", deps.UploadsController.Complete)

			api.GET("/events", controllers.Events(deps.EventsHub))
		}
	}
}
