package gnudb

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/b0bbywan/go-cd-cuer/types"
	"io"
	"net/http"
	"strings"
)

const (
	gnudbURL      = "http://yw443mcz.gnudb.org/~cddb/cddb.cgi"
	gnuHello      = "nas@bobbywan.me+bobbywan.me+rasponkyo+0.1"
	// Response keys for GNUDB
	keyTitle = "DTITLE="
	keyYear  = "DYEAR="
	keyGenre = "DGENRE="
	keyTrack = "TTITLE"

)

func FetchDiscInfo(discID string) (*types.DiscInfo, error) {
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
