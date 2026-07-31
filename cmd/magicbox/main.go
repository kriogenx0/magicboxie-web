package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"

	"magicbox/internal/auth"
	"magicbox/internal/config"
	"magicbox/internal/controllers"
	"magicbox/internal/db"
	"magicbox/internal/routes"
	"magicbox/internal/services/events"
	"magicbox/internal/services/library"
	"magicbox/internal/services/music"
	"magicbox/internal/services/tmdb"
	"magicbox/internal/services/transcode"
	"magicbox/internal/services/upload"
	"magicbox/internal/web"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash-password" {
		runHashPassword()
		return
	}

	defaultConfigPath := os.Getenv("MAGICBOX_CONFIG")
	if defaultConfigPath == "" {
		defaultConfigPath = "/etc/magicbox/config.yaml"
	}
	configPath := flag.String("config", defaultConfigPath, "path to config YAML file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("magicbox: %v", err)
	}

	for _, dir := range []string{cfg.DataDir, cfg.MoviesDir, cfg.MusicDir} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("magicbox: creating directory %q: %v", dir, err)
		}
	}

	for _, bin := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(bin); err != nil {
			log.Fatalf("magicbox: required dependency %q not found on PATH: %v", bin, err)
		}
	}

	gormDB, err := db.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("magicbox: %v", err)
	}

	authManager, err := auth.NewManager(cfg.DataDir)
	if err != nil {
		log.Fatalf("magicbox: %v", err)
	}

	tmdbClient := tmdb.NewClient(cfg.TMDB.APIReadToken)
	importer := library.NewImporter(gormDB, cfg.MoviesDir, cfg.DataDir, tmdbClient)
	musicImporter := music.NewImporter(gormDB, cfg.MusicDir, cfg.DataDir)

	eventsHub := events.NewHub()
	transcodeManager := transcode.NewManager(gormDB, cfg.MoviesDir, cfg.Transcode.Preset, cfg.Transcode.CRF, cfg.Transcode.MaxConcurrentJobs, eventsHub)
	importer.OnNeedsTranscode = transcodeManager.Enqueue
	transcodeManager.Start(context.Background())

	authController := controllers.NewAuthController(cfg, authManager)
	itemsController := controllers.NewItemsController(gormDB, importer, musicImporter, cfg.MoviesDir, cfg.DataDir)
	videosController := controllers.NewVideosController(gormDB, cfg.MoviesDir)
	audioController := controllers.NewAudioController(gormDB, cfg.MusicDir)

	uploadManager, err := upload.NewManager(gormDB, cfg.MoviesDir)
	if err != nil {
		log.Fatalf("magicbox: %v", err)
	}
	uploadsController := controllers.NewUploadsController(uploadManager, cfg.MoviesDir, cfg.MusicDir, importer, musicImporter)

	router := gin.Default()
	routes.Register(router, routes.Dependencies{
		AuthManager:       authManager,
		AuthController:    authController,
		ItemsController:   itemsController,
		VideosController:  videosController,
		AudioController:   audioController,
		UploadsController: uploadsController,
		EventsHub:         eventsHub,
	})

	fsys, err := web.FileSystem()
	if err != nil {
		log.Fatalf("magicbox: %v", err)
	}
	web.RegisterSPA(router, fsys)

	log.Printf("magicbox: listening on %s", cfg.ListenAddr)
	if err := router.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("magicbox: server error: %v", err)
	}
}

func runHashPassword() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: magicbox hash-password <password>")
		os.Exit(1)
	}
	hash, err := auth.HashPassword(os.Args[2])
	if err != nil {
		log.Fatalf("magicbox: %v", err)
	}
	fmt.Println(hash)
}
