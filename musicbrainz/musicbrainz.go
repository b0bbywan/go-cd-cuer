package musicbrainz

import (
	"errors"
	"fmt"
	"encoding/json"
	"net/http"

	"github.com/b0bbywan/go-disc-cuer/types"
)

const (
	mbURL         = "https://musicbrainz.org/ws/2"
)

func FetchReleaseByID(releaseID string) (*types.DiscInfo, error) {
	url := fmt.Sprintf("%s/release/%s?inc=artists+recordings&fmt=json", mbURL, releaseID)
	var release types.MBRelease
	if err := fetchJSON(url, &release); err != nil {
		return nil, err
	}
	return convertReleaseToDiscInfo(release)

}

func FetchReleaseByToc(mbToc string) (*types.DiscInfo, error) {
	url := fmt.Sprintf("%s/discid/-?toc=%s&inc=artists+recordings&fmt=json", mbURL, mbToc)
	var result types.ReleaseResult
	if err := fetchJSON(url, &result); err != nil {
		return nil, err
	}

	if len(result.Releases) == 0 {
		return nil, errors.New("no release data found")
	}

	release := result.Releases[0]
	return convertReleaseToDiscInfo(release)
}

func convertReleaseToDiscInfo(release types.MBRelease) (*types.DiscInfo, error) {
	tracks := make([]string, len(release.Media[0].Tracks))
	for i, track := range release.Media[0].Tracks {
		tracks[i] = track.Title
	}

	return &types.DiscInfo{
		ID:          release.ID,
		Title:       release.Title,
		Artist:      release.ArtistCredit[0].Name,
		ReleaseDate: release.Date,
		Tracks:      tracks,
	}, nil
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
