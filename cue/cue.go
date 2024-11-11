package cue

import (
	"fmt"
	"github.com/b0bbywan/go-cd-cuer/musicbrainz"
	"github.com/b0bbywan/go-cd-cuer/types"
	"github.com/b0bbywan/go-cd-cuer/utils"
	"log"
	"os"
	"path/filepath"
)

func Generate(providedDiscID, musicbrainzID string, overwrite bool) (string, error) {
	if err := utils.RemoveEnvFile(); err != nil {
//		log.Fatalf("error removing env file: %v", err)
		return "", err
	}

	discInfo, discID, err := fetchDiscInfoFromFlags(providedDiscID, musicbrainzID)
	if err != nil {
//		log.Fatalf("error parsing options: %v", err)
		return "", err
	}

	var gnuToc string
	if discID == "" {
		if gnuToc, discID, err = utils.GetTocAndDiscID(); err != nil {
//			log.Fatalf("error retrieving disc ID: %v", err)
			return "", err
		}
	}
	cueFilePath := utils.CachePlaylistPath(discID)

	if utils.CheckIfPlaylistExists(cueFilePath) && !overwrite {
		return cueFilePath, nil
	}

	if discInfo != nil && discID != "" {
		return finalizeIfSuccess(discInfo, cueFilePath)
	}
	var mbToc string
	if mbToc, err = utils.GetMusicBrainzDiscIDFromCmd(); err != nil {
		return "", err
//		log.Fatalf("error retrieving MusicBrainz disc ID: %v", err)
	}

	if err = utils.CreateFolderIfNeeded(cueFilePath); err != nil {
		return "", err
//		log.Fatalf("error creating folder for discID: %v", err)
	}

	// Fetch DiscInfo concurrently
	if discInfo, err = fetchDiscInfoConcurrently(gnuToc, mbToc); err != nil {
		return "", err
//		log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}

	return finalizeIfSuccess(discInfo, cueFilePath)
}

// fetchDiscInfoFromFlags returns DiscInfo, disc ID, and an error based on provided options.
func fetchDiscInfoFromFlags(musicbrainzID, providedDiscID string) (*types.DiscInfo, string, error) {
	// Enforce --musicbrainz with --disc-id
	if providedDiscID != "" && musicbrainzID == "" {
		return nil, "", fmt.Errorf("error: --disc-id option requires --musicbrainz to be set")
	}

	// If --musicbrainz is provided, fetch DiscInfo directly from MusicBrainz
	if musicbrainzID != "" {
		discInfo, err := musicbrainz.FetchReleaseByID(musicbrainzID)
		if err != nil {
			return nil, "", err
		}
		return discInfo, providedDiscID, nil
	}
	return nil, "", nil
}

func finalizeIfSuccess(discInfo *types.DiscInfo, cueFilePath string) (string, error) {
    if err := fetchCoverArtIfNeeded(discInfo, cueFilePath); err != nil {
        log.Printf("Error fetching cover art: %v", err)
    }
	// Generate the CUE file and save
	if err := generateCueFile(discInfo, cueFilePath); err != nil {
		return "", err
//		log.Fatalf("error: failed to generate CUE file: %v", err)
	}
	utils.SaveEnvFile(cueFilePath)
	log.Printf("info: Playlist generated at %s", cueFilePath)
	return cueFilePath, nil
}

func generateCueFile(info *types.DiscInfo, cueFilePath string) error {
	file, err := os.Create(cueFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if info.CoverArtPath == "" {
		coverFilePath := utils.CacheCoverArtPath(filepath.Base(filepath.Dir(cueFilePath)))
		if err := fetchCoverArt(info.ID, coverFilePath); err == nil {
			info.CoverArtPath = coverFilePath
		} else {
			log.Printf("error getting cover: %v", err)
		}
	}

	var content string
	if info.ReleaseDate != "" {
		content += fmt.Sprintf("REM DATE \"%s\"\n", info.ReleaseDate)
	}
	if info.Genre != "" {
		content += fmt.Sprintf("REM GENRE \"%s\"\n", info.Genre)
	}
	if info.CoverArtPath != "" {
		content += fmt.Sprintf("REM COVER \"%s\"\n", info.CoverArtPath)
	}
	content += fmt.Sprintf("PERFORMER \"%s\"\nTITLE \"%s\"\n", info.Artist, info.Title)

	for i, track := range info.Tracks {
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n",
			i+1, i+1, track)
	}

	_, err = file.WriteString(content)
	return err
}
