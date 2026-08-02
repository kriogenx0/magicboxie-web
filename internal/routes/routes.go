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
//     re-match), namespaced under /magicbox/* so they're clearly separate
//     from the standard surface.
func Register(router *gin.Engine, deps Dependencies) {
	router.GET("/System/Info/Public", deps.AuthController.SystemInfoPublic)
	router.POST("/Users/AuthenticateByName", deps.AuthController.AuthenticateByName)

	// Images are unauthenticated, matching real Jellyfin's convention (so
	// <img> tags never need a token).
	router.GET("/Items/:itemId/Images/Primary", deps.ItemsController.PrimaryImage)
	router.GET("/Items/:itemId/Images/Backdrop/:index", deps.ItemsController.BackdropImage)

	router.GET("/magicbox/health", controllers.Health)

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

		magicbox := authorized.Group("/magicbox")
		{
			magicbox.POST("/library/scan", deps.ItemsController.Scan)
			magicbox.POST("/library/music/scan", deps.ItemsController.MusicScan)
			magicbox.POST("/items/:itemId/match", deps.ItemsController.Match)
			magicbox.GET("/items/search", deps.ItemsController.Search)

			magicbox.POST("/uploads", deps.UploadsController.Create)
			magicbox.GET("/uploads/:id", deps.UploadsController.Status)
			magicbox.PUT("/uploads/:id/chunk", deps.UploadsController.Chunk)
			magicbox.POST("/uploads/:id/complete", deps.UploadsController.Complete)

			magicbox.GET("/events", controllers.Events(deps.EventsHub))
		}
	}
}
