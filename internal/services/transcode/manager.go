// Package transcode runs background ffmpeg transcode jobs for movies that
// aren't already browser/iOS compatible: a small in-process worker pool
// (no external broker needed for a single-box personal server), with DB-
// backed job state so progress survives a server restart.
package transcode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"magicbox/internal/models"
	"magicbox/internal/services/events"
)

const progressUpdateInterval = 1 * time.Second

type Manager struct {
	db            *gorm.DB
	moviesDir     string
	preset        string
	crf           int
	maxConcurrent int
	hub           *events.Hub

	queue chan uint // movie IDs awaiting a worker
}

func NewManager(db *gorm.DB, moviesDir, preset string, crf, maxConcurrent int, hub *events.Hub) *Manager {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Manager{
		db:            db,
		moviesDir:     moviesDir,
		preset:        preset,
		crf:           crf,
		maxConcurrent: maxConcurrent,
		hub:           hub,
		queue:         make(chan uint, 256),
	}
}

// Start launches the worker pool and recovers any jobs left queued/running
// by a previous process (a "running" job means the server crashed mid
// encode; it's reset to queued and re-run from scratch).
func (m *Manager) Start(ctx context.Context) {
	var staleJobs []models.Job
	m.db.Where("status IN ?", []string{models.JobStatusQueued, models.JobStatusRunning}).Find(&staleJobs)
	for _, job := range staleJobs {
		if job.Status == models.JobStatusRunning {
			m.db.Model(&models.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
				"status":           models.JobStatusQueued,
				"progress_percent": 0,
			})
		}
	}

	// Movies that ended up needs_transcode without an active job (e.g.
	// imported before this manager was wired up) get swept in too.
	var orphaned []models.Movie
	m.db.Where("status = ?", models.MovieStatusNeedsTranscode).Find(&orphaned)
	for _, movie := range orphaned {
		var count int64
		m.db.Model(&models.Job{}).
			Where("movie_id = ? AND type = ? AND status IN ?", movie.ID, models.JobTypeTranscode, []string{models.JobStatusQueued, models.JobStatusRunning}).
			Count(&count)
		if count == 0 {
			m.Enqueue(movie.ID)
		}
	}

	for i := 0; i < m.maxConcurrent; i++ {
		go m.worker(ctx)
	}

	for _, job := range staleJobs {
		m.queue <- job.MovieID
	}
}

// Enqueue creates a queued transcode job for movieID and schedules it.
func (m *Manager) Enqueue(movieID uint) {
	job := &models.Job{MovieID: movieID, Type: models.JobTypeTranscode, Status: models.JobStatusQueued}
	if err := m.db.Create(job).Error; err != nil {
		log.Printf("transcode: failed to create job for movie %d: %v", movieID, err)
		return
	}
	m.queue <- movieID
}

func (m *Manager) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case movieID := <-m.queue:
			m.process(ctx, movieID)
		}
	}
}

func (m *Manager) process(ctx context.Context, movieID uint) {
	var movie models.Movie
	if err := m.db.First(&movie, movieID).Error; err != nil {
		log.Printf("transcode: movie %d not found: %v", movieID, err)
		return
	}

	var job models.Job
	if err := m.db.Where("movie_id = ? AND type = ? AND status = ?", movieID, models.JobTypeTranscode, models.JobStatusQueued).
		Order("created_at desc").First(&job).Error; err != nil {
		log.Printf("transcode: no queued job found for movie %d: %v", movieID, err)
		return
	}

	now := time.Now()
	m.db.Model(&job).Updates(map[string]interface{}{"status": models.JobStatusRunning, "started_at": now})
	m.db.Model(&movie).Update("status", models.MovieStatusTranscoding)

	srcPath := filepath.Join(m.moviesDir, movie.PlayableRelpath)
	tmpDir := filepath.Join(m.moviesDir, ".magicbox", "transcode-tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		m.fail(&job, &movie, fmt.Errorf("creating transcode tmp dir: %w", err))
		return
	}
	tmpOutPath := filepath.Join(tmpDir, fmt.Sprintf("%d.mp4", movie.ID))

	if err := m.runFFmpeg(ctx, &job, &movie, srcPath, tmpOutPath); err != nil {
		os.Remove(tmpOutPath)
		m.fail(&job, &movie, err)
		return
	}

	finalRelpath, err := m.finalize(srcPath, tmpOutPath, movie.SourceRelpath)
	if err != nil {
		m.fail(&job, &movie, fmt.Errorf("finalizing transcode output: %w", err))
		return
	}

	stat, statErr := os.Stat(filepath.Join(m.moviesDir, finalRelpath))

	updates := map[string]interface{}{
		"status":           models.MovieStatusReady,
		"playable_relpath": finalRelpath,
		"video_codec":      "h264",
		"audio_codec":      "aac",
		"container":        "mov,mp4,m4a,3gp,3g2,mj2",
		"error_message":    "",
	}
	if statErr == nil {
		updates["file_size_bytes"] = stat.Size()
	}
	m.db.Model(&movie).Updates(updates)

	finishedAt := time.Now()
	m.db.Model(&job).Updates(map[string]interface{}{
		"status":           models.JobStatusCompleted,
		"progress_percent": 100,
		"finished_at":      finishedAt,
	})

	m.hub.Broadcast(events.Event{Type: "job_completed", Data: eventData{
		"movie_id": movie.ID,
		"job_id":   job.ID,
		"status":   models.MovieStatusReady,
	}})
}

// finalize deletes the (now-superseded) original source file -- per the
// single-file retention model, the transcoded copy is the only file that
// remains -- and moves the transcoded output into place next to where the
// original lived, deduping the filename if needed.
func (m *Manager) finalize(srcPath, tmpOutPath, sourceRelpath string) (finalRelpath string, err error) {
	destDir := filepath.Dir(filepath.Join(m.moviesDir, sourceRelpath))
	base := strings.TrimSuffix(filepath.Base(sourceRelpath), filepath.Ext(sourceRelpath))

	destPath := filepath.Join(destDir, base+".mp4")
	for i := 2; destPath == srcPath; i++ {
		// Extremely rare: transcoded name collides with the still-present
		// source path (e.g. source was already "Title.mp4" but incompatible
		// for codec/resolution reasons). Disambiguate before deleting src.
		destPath = filepath.Join(destDir, fmt.Sprintf("%s (%d).mp4", base, i))
	}

	if err := os.Remove(srcPath); err != nil {
		return "", fmt.Errorf("deleting original source file: %w", err)
	}
	if err := os.Rename(tmpOutPath, destPath); err != nil {
		return "", fmt.Errorf("moving transcoded output into place: %w", err)
	}

	rel, err := filepath.Rel(m.moviesDir, destPath)
	if err != nil {
		return "", err
	}
	return rel, nil
}

func (m *Manager) fail(job *models.Job, movie *models.Movie, cause error) {
	log.Printf("transcode: movie %d failed: %v", movie.ID, cause)
	now := time.Now()
	m.db.Model(job).Updates(map[string]interface{}{
		"status":      models.JobStatusFailed,
		"log_tail":    truncate(cause.Error(), 4000),
		"finished_at": now,
	})
	m.db.Model(movie).Updates(map[string]interface{}{
		"status":        models.MovieStatusError,
		"error_message": cause.Error(),
	})
	m.hub.Broadcast(events.Event{Type: "job_failed", Data: eventData{
		"movie_id": movie.ID,
		"job_id":   job.ID,
		"error":    cause.Error(),
	}})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// eventData is a plain map alias for SSE event payloads, keeping this
// services package free of any dependency on the HTTP/controller layer.
type eventData = map[string]interface{}

func (m *Manager) runFFmpeg(ctx context.Context, job *models.Job, movie *models.Movie, srcPath, outPath string) error {
	args := []string{
		"-y",
		"-i", srcPath,
		"-vf", "scale=-2:'min(1080,ih)'",
		"-c:v", "libx264",
		"-preset", m.preset,
		"-crf", strconv.Itoa(m.crf),
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "+faststart",
		"-progress", "pipe:1",
		"-nostats",
		outPath,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrTail strings.Builder
	cmd.Stderr = &tailWriter{limit: 4000, builder: &stderrTail}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	m.watchProgress(job, movie, stdout)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, stderrTail.String())
	}
	return nil
}

// watchProgress reads ffmpeg's `-progress pipe:1` key=value stream and
// updates the job's progress, throttled to roughly once per second so we
// don't hammer SQLite (or the SSE clients) on every frame.
func (m *Manager) watchProgress(job *models.Job, movie *models.Movie, stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	lastUpdate := time.Time{}

	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if key == "out_time" && movie.DurationSeconds > 0 {
			if elapsed, ok := parseFFmpegTime(value); ok {
				percent := (elapsed / movie.DurationSeconds) * 100
				if percent > 100 {
					percent = 100
				}
				if percent < 0 {
					percent = 0
				}

				if time.Since(lastUpdate) >= progressUpdateInterval {
					lastUpdate = time.Now()
					m.db.Model(job).Update("progress_percent", percent)
					m.hub.Broadcast(events.Event{Type: "job_progress", Data: eventData{
						"movie_id":         movie.ID,
						"job_id":           job.ID,
						"progress_percent": percent,
					}})
				}
			}
		}
	}
}

// parseFFmpegTime parses ffmpeg's "-progress" out_time value (HH:MM:SS.ffffff)
// into total seconds.
func parseFFmpegTime(s string) (float64, bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, false
	}
	hours, err1 := strconv.ParseFloat(parts[0], 64)
	minutes, err2 := strconv.ParseFloat(parts[1], 64)
	seconds, err3 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, false
	}
	return hours*3600 + minutes*60 + seconds, true
}

// tailWriter keeps only the last `limit` bytes written to it, for capturing
// a bounded ffmpeg stderr tail on failure.
type tailWriter struct {
	limit   int
	builder *strings.Builder
}

func (w *tailWriter) Write(p []byte) (int, error) {
	w.builder.Write(p)
	if w.builder.Len() > w.limit {
		s := w.builder.String()
		w.builder.Reset()
		w.builder.WriteString(s[len(s)-w.limit:])
	}
	return len(p), nil
}
