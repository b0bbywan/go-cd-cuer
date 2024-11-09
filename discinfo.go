package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	gnudbURL      = "http://yw443mcz.gnudb.org/~cddb/cddb.cgi"
	gnuHello      = "nas@bobbywan.me+bobbywan.me+rasponkyo+0.1"
	mbURL         = "https://musicbrainz.org/ws/2"
	coverArtURL   = "https://coverartarchive.org/release"
)

const (
	// Response keys for GNUDB
	keyTitle = "DTITLE="
	keyYear  = "DYEAR="
	keyGenre = "DGENRE="
	keyTrack = "TTITLE"
)

func fetchGNUDBDiscInfo(discID string) (*DiscInfo, error) {
	client := &http.Client{}
	// First, query GNDB for a match
	queryURL := fmt.Sprintf("%s?cmd=cddb+query+%s&hello=%s&proto=6", gnudbURL, discID, gnuHello)
	resp, err := makeGnuRequest(client, queryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Found exact matches") {
		return nil, errors.New("no exact match found in GNUDB")
	}

	// Extract GNDB ID and title from query response
	lines := strings.Split(string(body), "\n")
	gnudbID := strings.Fields(lines[1])[1]

	// Now read full metadata
	readURL := fmt.Sprintf("%s?cmd=cddb+read+data+%s&hello=%s&proto=6", gnudbURL, gnudbID, gnuHello)
	resp, err = makeGnuRequest(client, readURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseGNUDBResponse(resp.Body)
}

func parseGNUDBResponse(body io.Reader) (*DiscInfo, error) {
	scanner := bufio.NewScanner(body)
	discInfo := &DiscInfo{}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, keyTitle):
			titleLine := strings.TrimPrefix(line, keyTitle)
			parts := strings.SplitN(titleLine, " / ", 2)
			if len(parts) == 2 {
				discInfo.Artist, discInfo.Title = parts[0], parts[1]
			}
		case strings.HasPrefix(line, keyYear):
			discInfo.ReleaseDate = strings.TrimPrefix(line, keyYear)
		case strings.HasPrefix(line, keyGenre):
			discInfo.Genre = strings.TrimPrefix(line, keyGenre)
		case strings.HasPrefix(line, keyTrack):
			track := strings.TrimPrefix(line, keyTrack)
			discInfo.Tracks = append(discInfo.Tracks, track)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if discInfo.Title == "" {
		return nil, errors.New("error: no valid title in GNUDB data")
	}

	return discInfo, nil
}

func fetchMusicBrainzRelease(discID string) (*DiscInfo, error) {
	url := fmt.Sprintf("%s/discid/-?toc=%s&inc=artists+recordings&fmt=json", mbURL, discID)
	var result ReleaseResult
	if err := fetchJSON(url, &result); err != nil {
		return nil, err
	}

	if len(result.Releases) == 0 {
		return nil, errors.New("no release data found")
	}

	release := result.Releases[0]
	tracks := make([]string, len(release.Media[0].Tracks))
	for i, track := range release.Media[0].Tracks {
		tracks[i] = track.Title
	}

	return &DiscInfo{
		ID:          release.ID,
		Title:       release.Title,
		Artist:      release.ArtistCredit[0].Name,
		ReleaseDate: release.Date,
		Tracks:      tracks,
	}, nil
}

// Function to fetch disc info from both services using goroutines and WaitGroup
func fetchDiscInfoConcurrently(discID, mbToc string) (*DiscInfo, error) {
	var wg sync.WaitGroup
	var gndbDiscInfo, mbDiscInfo *DiscInfo
	var gndbErr, mbErr error

	wg.Add(2)

	// Fetch from GNUDB
	go func() {
		defer wg.Done()
		gndbDiscInfo, gndbErr = fetchGNUDBDiscInfo(discID)
	}()

	// Fetch from MusicBrainz
	go func() {
		defer wg.Done()
		mbDiscInfo, mbErr = fetchMusicBrainzRelease(mbToc)
	}()

	// Wait for both fetches to complete
	wg.Wait()

	return selectDiscInfo(gndbDiscInfo, gndbErr, mbDiscInfo, mbErr)
}

func selectDiscInfo(gndbDiscInfo *DiscInfo, gndbErr error, mbDiscInfo *DiscInfo, mbErr error) (*DiscInfo, error) {

	// Decide on the final discInfo, prioritizing GNUDB data where available
	finalDiscInfo := &DiscInfo{}
	if gndbErr == nil {
		*finalDiscInfo = *gndbDiscInfo
	} else if mbErr == nil {
		*finalDiscInfo = *mbDiscInfo
	}

	// Use MusicBrainz ID regardless of source priority
	if mbDiscInfo != nil {
		finalDiscInfo.ID = mbDiscInfo.ID
	}

	// If both failed, return an error
	if gndbErr != nil && mbErr != nil {
		return nil, fmt.Errorf("failed to fetch from both sources: GNUDB error: %v; MusicBrainz error: %v", gndbErr, mbErr)
	}

	return finalDiscInfo, nil
}

func fetchCoverArt(mbID, coverFile string) error {
	url := fmt.Sprintf("%s/%s/front", coverArtURL, mbID)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(coverFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
