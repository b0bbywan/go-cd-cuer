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

// FetchDiscInfo queries GNUDB to retrieve disc metadata based on the disc's table of contents (TOC).
//
// Parameters:
//   - gnuToc (string): The disc's TOC, formatted for GNUDB queries.
//
// Returns:
//   - *types.DiscInfo: A struct containing the disc's metadata (artist, title, tracks, etc.).
//   - error: An error if the query or data retrieval fails, or if `gnuHelloEmail` is not configured.
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

// queryGNUDB performs the initial query to GNUDB to find a matching record for the given TOC.
//
// Parameters:
//   - client (*http.Client): HTTP client for making requests.
//   - gnuToc (string): The disc's TOC, formatted for GNUDB queries.
//
// Returns:
//   - string: The GNUDB ID of the matching record.
//   - error: An error if the query fails, the response cannot be read, or no match is found.
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

// extractGnuDBID extracts the GNUDB ID from a successful query response.
//
// Parameters:
//   - response (string): The raw response string from the GNUDB query.
//
// Returns:
//   - string: The extracted GNUDB ID.
//   - error: An error if the response format is invalid.
func extractGnuDBID(response string) (string, error) {
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return "", errors.New("invalid GNUDB response format")
	}
	return strings.Fields(lines[1])[1], nil
}

// fetchFullMetadata retrieves detailed disc metadata from GNUDB using the record's ID.
//
// Parameters:
//   - client (*http.Client): HTTP client for making requests.
//   - gnudbID (string): The ID of the record in GNUDB.
//
// Returns:
//   - *types.DiscInfo: A struct containing the disc's metadata (artist, title, tracks, etc.).
//   - error: An error if the metadata cannot be retrieved or parsed.
func fetchFullMetadata(client *http.Client, gnudbID string) (*types.DiscInfo, error) {
	readURL := fmt.Sprintf("%s?cmd=cddb+read+data+%s&hello=%s&proto=6", gnudbURL, gnudbID, gnuHello)
	resp, err := makeGnuRequest(client, readURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseGNUDBResponse(resp.Body)
}

// parseGNUDBResponse parses the response from a detailed metadata request to GNUDB.
//
// Parameters:
//   - body (io.Reader): The response body from the GNUDB read command.
//
// Returns:
//   - *types.DiscInfo: A struct containing the parsed disc metadata.
//   - error: An error if parsing fails or if required fields (e.g., title) are missing.
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

// makeGnuRequest performs an HTTP GET request with a predefined User-Agent header.
//
// Parameters:
//   - client (*http.Client): HTTP client for making requests.
//   - url (string): The URL to send the GET request to.
//
// Returns:
//   - *http.Response: The HTTP response object.
//   - error: An error if the request fails.
func makeGnuRequest(client *http.Client, url string) (*http.Response, error) {
	userAgent := "curl/8.9.1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	return client.Do(req)
}
