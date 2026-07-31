package library

import "testing"

func TestParseFilename(t *testing.T) {
	cases := []struct {
		filename  string
		wantTitle string
		wantYear  int
	}{
		{"Movie.Title.2015.1080p.BluRay.x264-GROUP.mkv", "Movie Title", 2015},
		{"Movie Title (2015).mkv", "Movie Title", 2015},
		{"Some.Movie.mkv", "Some Movie", 0},
		{"The.Matrix.1999.2160p.UHD.HDR.x265-GROUP.mp4", "The Matrix", 1999},
		{"Spirited_Away_2001_WEBRip.mkv", "Spirited Away", 2001},
		{"Untagged Movie.avi", "Untagged Movie", 0},
		{"Some.Movie.2160p.mkv", "Some Movie", 0},
	}

	for _, tc := range cases {
		gotTitle, gotYear := ParseFilename(tc.filename)
		if gotTitle != tc.wantTitle || gotYear != tc.wantYear {
			t.Errorf("ParseFilename(%q) = (%q, %d), want (%q, %d)",
				tc.filename, gotTitle, gotYear, tc.wantTitle, tc.wantYear)
		}
	}
}
