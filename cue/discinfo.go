package cue

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/b0bbywan/go-disc-cuer/gnudb"
	"github.com/b0bbywan/go-disc-cuer/musicbrainz"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"

)

const (
	coverArtURL   = "https://coverartarchive.org/release"
)

// FetchCoverArtIfNeeded checks if cover art is missing and fetches it if necessary.
func fetchCoverArtIfNeeded(discInfo *types.DiscInfo, cueFilePath string) error {
    if discInfo.CoverArtPath == "" {
        coverFilePath := utils.CacheCoverArtPath(filepath.Base(filepath.Dir(cueFilePath)))
        if err := fetchCoverArt(discInfo.ID, coverFilePath); err == nil {
            discInfo.CoverArtPath = coverFilePath
        } else {
            return fmt.Errorf("error getting cover: %v", err)
        }
    }
    return nil
}

func fetchCoverArt(mbID, coverFile string) error {
	url := fmt.Sprintf("%s/%s/front", coverArtURL, mbID)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to fetch cover art: received status code %d", resp.StatusCode)
	}

	file, err := os.Create(coverFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// Function to fetch disc info from both services using goroutines and WaitGroup
func fetchDiscInfoConcurrently(gnuToc, mbToc string) (*types.DiscInfo, error) {
	var wg sync.WaitGroup
	var gndbDiscInfo, mbDiscInfo *types.DiscInfo
	var gndbErr, mbErr error
	formattedGnuTOC := strings.ReplaceAll(gnuToc, " ", "+")
	formattedMBTOC := strings.ReplaceAll(mbToc, " ", "+")

	wg.Add(2)

	// Fetch from GNUDB
	go func() {
		defer wg.Done()
		gndbDiscInfo, gndbErr = gnudb.FetchDiscInfo(formattedGnuTOC)
	}()

	// Fetch from MusicBrainz
	go func() {
		defer wg.Done()
		mbDiscInfo, mbErr = musicbrainz.FetchReleaseByToc(formattedMBTOC)
	}()

	// Wait for both fetches to complete
	wg.Wait()

	return selectDiscInfo(gndbDiscInfo, gndbErr, mbDiscInfo, mbErr)
}

func selectDiscInfo(gndbDiscInfo *types.DiscInfo, gndbErr error, mbDiscInfo *types.DiscInfo, mbErr error) (*types.DiscInfo, error) {

	// Decide on the final discInfo, prioritizing GNUDB data where available
	finalDiscInfo := &types.DiscInfo{}
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
