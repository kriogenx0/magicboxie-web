// Package tmdb is a minimal client for the three TMDB v3 API endpoints
// MagicBoxie needs: search, movie details (with credits), and image download.
package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL  = "https://api.themoviedb.org/3"
	imageURL = "https://image.tmdb.org/t/p"
)

// ErrNotConfigured is returned when no API read token is set. Callers should
// treat this the same as "no match found".
var ErrNotConfigured = errors.New("tmdb: api_read_token not configured")

type Client struct {
	readToken string
	http      *http.Client
	limiter   *rate.Limiter
}

func NewClient(readToken string) *Client {
	return &Client{
		readToken: readToken,
		http:      &http.Client{Timeout: 15 * time.Second},
		limiter:   rate.NewLimiter(4, 4), // conservative: 4 req/s, burst 4
	}
}

type SearchResult struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	ReleaseDate  string  `json:"release_date"`
	Popularity   float64 `json:"popularity"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
}

type searchResponse struct {
	Results []SearchResult `json:"results"`
}

// SearchMovie finds candidate matches by title (+ optional year), returning
// the most popular result first.
func (c *Client) SearchMovie(ctx context.Context, title string, year int) ([]SearchResult, error) {
	if c.readToken == "" {
		return nil, ErrNotConfigured
	}

	q := fmt.Sprintf("%s/search/movie?query=%s", baseURL, url.QueryEscape(title))
	if year > 0 {
		q += fmt.Sprintf("&primary_release_year=%d", year)
	}

	var resp searchResponse
	if err := c.getJSON(ctx, q, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

type CastMember struct {
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
}

type MovieDetails struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Overview string `json:"overview"`
	Runtime  int    `json:"runtime"`
	Genres   []struct {
		Name string `json:"name"`
	} `json:"genres"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	Credits      struct {
		Cast []CastMember `json:"cast"`
	} `json:"credits"`
}

// GetMovieDetails fetches full details + credits in one round trip via
// append_to_response.
func (c *Client) GetMovieDetails(ctx context.Context, tmdbID int) (*MovieDetails, error) {
	if c.readToken == "" {
		return nil, ErrNotConfigured
	}

	q := fmt.Sprintf("%s/movie/%d?append_to_response=credits", baseURL, tmdbID)

	var details MovieDetails
	if err := c.getJSON(ctx, q, &details); err != nil {
		return nil, err
	}
	return &details, nil
}

// DownloadImage fetches a poster/backdrop image (size e.g. "w780", "w1280").
func (c *Client) DownloadImage(ctx context.Context, size, imagePath string) ([]byte, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s%s", imageURL, size, imagePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tmdb: image download %q failed: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) getJSON(ctx context.Context, url string, dest interface{}) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.readToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			backoff := time.Duration(attempt+1) * time.Second
			lastErr = fmt.Errorf("tmdb: rate limited (429)")
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("tmdb: request to %q failed: %s: %s", url, resp.Status, string(body))
		}
		return json.Unmarshal(body, dest)
	}
	return fmt.Errorf("tmdb: giving up after retries: %w", lastErr)
}
