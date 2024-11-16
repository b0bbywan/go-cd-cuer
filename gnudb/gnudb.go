package gnudb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/types"
)

const (
	// Response keys for GNUDB
	keyTitle = "DTITLE="
	keyYear  = "DYEAR="
	keyGenre = "DGENRE="
	keyTrack = "TTITLE"
)

var (
	gnuHello string
	gnudbURL string
)

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	gnuHello = fmt.Sprintf("%s+%s+%s+%s", config.GnuHelloEmail, hostname, config.AppName, config.AppVersion)
	gnudbURL = fmt.Sprintf("%s/~cddb/cddb.cgi", config.GnuDbUrl)
}

func FetchDiscInfo(gnuToc string) (*types.DiscInfo, error) {
	client := &http.Client{}

	if config.GnuHelloEmail == "" {
		return nil, fmt.Errorf("gnuHelloEmail is required in config.yaml or via environment variable to use gnuDB")
	}

	// First, query GNUDB for a match
	gnudbID, err := queryGNUDB(client, gnuToc)
	if err != nil {
		return nil, err
	}
	// Fetch the full metadata from GNDB
	discInfo, err := fetchFullMetadata(client, gnudbID)
	if err != nil {
		return nil, err
	}

	return discInfo, nil
}

func queryGNUDB(client *http.Client, gnuToc string) (string, error) {
	queryURL := fmt.Sprintf("%s?cmd=cddb+query+%s&hello=%s&proto=6", gnudbURL, gnuToc, gnuHello)
	resp, err := makeGnuRequest(client, queryURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if !strings.Contains(string(body), "Found exact matches") {
		return "", errors.New("no exact match found in GNUDB")
	}
	return extractGnuDBID(string(body))
}

func extractGnuDBID(response string) (string, error) {
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return "", errors.New("invalid GNUDB response format")
	}
	return strings.Fields(lines[1])[1], nil
}

func fetchFullMetadata(client *http.Client, gnudbID string) (*types.DiscInfo, error) {
	readURL := fmt.Sprintf("%s?cmd=cddb+read+data+%s&hello=%s&proto=6", gnudbURL, gnudbID, gnuHello)
	resp, err := makeGnuRequest(client, readURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseGNUDBResponse(resp.Body)
}

func parseGNUDBResponse(body io.Reader) (*types.DiscInfo, error) {
	scanner := bufio.NewScanner(body)
	discInfo := &types.DiscInfo{}

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
			discInfo.Tracks = append(discInfo.Tracks, track[1])
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

func makeGnuRequest(client *http.Client, url string) (*http.Response, error) {
	userAgent := "curl/8.9.1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	return client.Do(req)
}
