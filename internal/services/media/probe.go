// Package media wraps ffprobe for media inspection and defines the
// browser/iOS compatibility rule used to decide whether a file needs
// transcoding.
package media

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Info struct {
	Container       string
	DurationSeconds float64
	VideoCodec      string
	AudioCodec      string
	Width           int
	Height          int
	BitrateBps      int
}

type ffprobeFormat struct {
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	BitRate    string `json:"bit_rate"`
}

type ffprobeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type ffprobeOutput struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

// Probe runs ffprobe against a media file and extracts container, duration,
// and the primary video/audio codec + resolution.
func Probe(ctx context.Context, path string) (Info, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return Info{}, fmt.Errorf("ffprobe failed for %q: %w", path, err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return Info{}, fmt.Errorf("parsing ffprobe output for %q: %w", path, err)
	}

	info := Info{Container: parsed.Format.FormatName}
	if d, err := strconv.ParseFloat(parsed.Format.Duration, 64); err == nil {
		info.DurationSeconds = d
	}
	if b, err := strconv.Atoi(parsed.Format.BitRate); err == nil {
		info.BitrateBps = b
	}

	for _, s := range parsed.Streams {
		switch s.CodecType {
		case "video":
			if info.VideoCodec == "" {
				info.VideoCodec = s.CodecName
				info.Width = s.Width
				info.Height = s.Height
			}
		case "audio":
			if info.AudioCodec == "" {
				info.AudioCodec = s.CodecName
			}
		}
	}

	return info, nil
}

// IsBrowserCompatible reports whether a probed file can be played/streamed
// directly (mp4/mov container, h264/hevc video, aac audio, height <= 1080p)
// without needing a background transcode.
func (i Info) IsBrowserCompatible() bool {
	containerOK := strings.Contains(i.Container, "mp4") || strings.Contains(i.Container, "mov")
	videoOK := i.VideoCodec == "h264" || i.VideoCodec == "hevc"
	audioOK := i.AudioCodec == "aac"
	resolutionOK := i.Height == 0 || i.Height <= 1080
	return containerOK && videoOK && audioOK && resolutionOK
}
