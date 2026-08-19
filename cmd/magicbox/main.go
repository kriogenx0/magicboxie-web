package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/gin-gonic/gin"

	"magicbox/internal/auth"
	"magicbox/internal/config"
	"magicbox/internal/controllers"
	"magicbox/internal/db"
	"magicbox/internal/models"
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
	if len(os.Args) > 1 && os.Args[1] == "rematch-tmdb" {
		runRematchTMDB(os.Args[2:])
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
	videosController := controllers.NewVideosController(gormDB, cfg.MoviesDir, cfg.DataDir)
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

func runRematchTMDB(args []string) {
	flags := flag.NewFlagSet("rematch-tmdb", flag.ExitOnError)
	configPath := flags.String("config", "/etc/magicbox/config.yaml", "path to config YAML file")
	_ = flags.Parse(args)
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}
	client := tmdb.NewClient(cfg.TMDB.APIReadToken)
	importer := library.NewImporter(database, cfg.MoviesDir, cfg.DataDir, client)
	var movies []models.Movie
	if err := database.Order("id").Find(&movies).Error; err != nil {
		log.Fatal(err)
	}
	matched, unmatched, failed := 0, 0, 0
	for i := range movies {
		movie := &movies[i]
		results, err := importer.SearchTMDB(context.Background(), movie.Title, movie.Year)
		if err != nil {
			log.Printf("FAILED  %q: %v", movie.Title, err)
			failed++
			continue
		}
		if len(results) == 0 && movie.Year > 0 {
			results, err = importer.SearchTMDB(context.Background(), movie.Title, 0)
		}
		if err != nil || len(results) == 0 {
			log.Printf("NO MATCH %q", movie.Title)
			unmatched++
			continue
		}
		sort.Slice(results, func(i, j int) bool { return results[i].Popularity > results[j].Popularity })
		if err := importer.ApplyManualMatch(context.Background(), movie, results[0].ID); err != nil {
			log.Printf("FAILED  %q: %v", movie.Title, err)
			failed++
			continue
		}
		log.Printf("MATCHED %q -> TMDB %d", movie.Title, results[0].ID)
		matched++
	}
	fmt.Printf("TMDB refresh complete: %d matched, %d unmatched, %d failed\n", matched, unmatched, failed)
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
