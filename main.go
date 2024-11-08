package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	envFile       = "/tmp/cd_var.env"
	gnudbURL      = "http://yw443mcz.gnudb.org/~cddb/cddb.cgi"
	gnuHello      = "nas@bobbywan.me+bobbywan.me+rasponkyo+0.1"
	mbURL         = "https://musicbrainz.org/ws/2"
	coverArtURL   = "https://coverartarchive.org/release"
	cacheLocation = "/home/bobby/.cddb" // replace 'user' with appropriate user

	// Response keys for GNDB
	keyTitle = "DTITLE="
	keyYear  = "DYEAR="
	keyGenre = "DGENRE="
	keyTrack = "TTITLE"
)

type DiscInfo struct {
	ID           string
	Artist       string
	Title        string
	ReleaseDate  string
	Genre        string
	Tracks       []string
	CoverArtPath string
}

// Execute a shell command and return its output as a string
func runCommand(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: failed to fetch from URL %s, status code: %d", url, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func makeGnuRequest(client *http.Client, url, userAgent string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	return client.Do(req)
}

func parseDiscID(discID string) ([]int, int, error) {
	fields := strings.Fields(string(discID))
	if len(fields) < 3 {
		return nil, 0, fmt.Errorf("unexpected cd-discid output")
	}

	// Extract start sectors
	startSectors := make([]int, len(fields)-3)
	for i := 2; i < len(fields)-1; i++ {
		sector, err := strconv.Atoi(fields[i])
		if err != nil {
			return nil, 0, err
		}
		startSectors[i-2] = sector
	}

	// Extract last track length
	lastTrackLength, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return nil, 0, err
	}

	return startSectors, lastTrackLength, nil
}

func getMusicBrainzDiscID(discID string) (string, error) {
	startSectors, lastTrackLength, err := parseDiscID(discID)
	if err != nil {
		return "", err
	}
	numTracks := len(startSectors)
	totalSectors := startSectors[numTracks-1] + lastTrackLength

	var mbToc strings.Builder
	mbToc.WriteString(fmt.Sprintf("1 %d %d", numTracks, totalSectors))
	for _, sector := range startSectors {
		mbToc.WriteString(fmt.Sprintf(" %d", sector))
	}
	return mbToc.String(), nil
}

func getDiscID() (string, error) {
	return runCommand("cd-discid")
}

func getMusicBrainzDiscIDFromCmd() (string, error) {
	return runCommand("cd-discid", "--musicbrainz")
}

func fetchGNUDBDiscInfo(discID string) (*DiscInfo, error) {
	client := &http.Client{}
	// First, query GNDB for a match
	queryURL := fmt.Sprintf("%s?cmd=cddb+query+%s&hello=%s&proto=6", gnudbURL, discID, gnuHello)
	resp, err := makeGnuRequest(client, queryURL, "curl/8.9.1")
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
	resp, err = makeGnuRequest(client, readURL, "curl/8.9.1")
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
			track := strings.SplitN(line, "=", 2)
			if len(track) == 2 {
				discInfo.Tracks = append(discInfo.Tracks, track[1])
			}
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
	var result struct {
		Releases []struct {
			ID             string `json:"id"`
			Title          string `json:"title"`
			Date           string `json:"date"`
			ArtistCredit   []struct{ Name string } `json:"artist-credit"`
			Media          []struct {
				Tracks []struct{ Title string }
			} `json:"media"`
		} `json:"releases"`
	}
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

func generateCueFile(info *DiscInfo, cueFilePath string) error {
	file, err := os.Create(cueFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	content := fmt.Sprintf("REM DATE \"%s\"\nREM COVER \"%s\"\nPERFORMER \"%s\"\nTITLE \"%s\"\n",
		info.ReleaseDate, info.CoverArtPath, info.Artist, info.Title)

	for i, track := range info.Tracks {
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n",
			i+1, i+1, track)
	}

	_, err = file.WriteString(content)
	return err
}

func cachePlaylistPath(discID string) string {
	return filepath.Join(cacheLocation, fmt.Sprintf("%s.cue", discID))
}

func saveEnvFile(cueFile string) error {
	return os.WriteFile(envFile, []byte(fmt.Sprintf("CUE_FILE=%s", cueFile)), 0644)
}

func main() {
	if err := os.Remove(envFile); err != nil && !os.IsNotExist(err) {
		log.Fatalf("error removing env file: %v", err)
	}

	// Fetch disc ID
	gnuToc, err := getDiscID()
	if err != nil {
		log.Fatalf("error retrieving disc ID: %v", err)
	}
	log.Printf(gnuToc)
	discID := strings.Fields(gnuToc)[0]
//	mbToc, err := getMusicBrainzDiscID(gnuToc)
//	log.Printf(mbToc)
	mbToc, err := getMusicBrainzDiscIDFromCmd()
	if err != nil {
		log.Fatalf("error retrieving disc ID: %v", err)
	}
	log.Printf(mbToc)

	cueFilePath := cachePlaylistPath(discID)
	if _, err := os.Stat(cueFilePath); err == nil {
		saveEnvFile(cueFilePath)
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return
	}

	var discInfo *DiscInfo

	// Attempt to fetch from GNUDB
	discInfo, err = fetchGNUDBDiscInfo(strings.Replace(gnuToc, " ", "+", -1))
	if err != nil {
		log.Printf("info: GNUDB retrieval failed, trying MusicBrainz. Error: %v", err)

		// Attempt to fetch from MusicBrainz if GNUDB fails
		
		discInfo, err = fetchMusicBrainzRelease(strings.Replace(mbToc, " ", "+", -1))
		if err != nil {
			log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz")
		}
	}

	// If MusicBrainz, attempt cover fetch
	if discInfo.CoverArtPath == "" {
		coverFilePath := filepath.Join(cacheLocation, fmt.Sprintf("%s.jpg", discID))
		if fetchCoverArt(discInfo.ID, coverFilePath) == nil {
			discInfo.CoverArtPath = coverFilePath
		}
	}

	// Generate the CUE file and save
	if err := generateCueFile(discInfo, cueFilePath); err != nil {
		log.Fatalf("error: failed to generate CUE file: %v", err)
	}

	saveEnvFile(cueFilePath)
	log.Printf("info: Playlist generated at %s", cueFilePath)
}
