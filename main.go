package main

import (
	"flag"
	"fmt"
	"github.com/b0bbywan/go-cd-cuer/discinfo"
	"github.com/b0bbywan/go-cd-cuer/musicbrainz"
	"github.com/b0bbywan/go-cd-cuer/types"
	"github.com/b0bbywan/go-cd-cuer/utils"
	"log"
)

var (
	overwrite      bool
	musicbrainzID  string
	providedDiscID string

)

func init() {
	flag.BoolVar(&overwrite, "overwrite", false, "force regenerating the CUE file even if it exists")
	flag.StringVar(&musicbrainzID, "musicbrainz", "", "specify MusicBrainz release ID directly")
	flag.StringVar(&providedDiscID, "disc-id", "", "specify disc ID directly")
}

// fetchDiscInfoFromFlags returns DiscInfo, disc ID, and an error based on provided options.
func fetchDiscInfoFromFlags() (*types.DiscInfo, string, error) {
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

func finalizeIfSuccess(discInfo *types.DiscInfo, cueFilePath string) {
	// Generate the CUE file and save
	if err := utils.GenerateCueFile(discInfo, cueFilePath); err != nil {
		log.Fatalf("error: failed to generate CUE file: %v", err)
	}
	utils.SaveEnvFile(cueFilePath)
	log.Printf("info: Playlist generated at %s", cueFilePath)
}

func main() {
	flag.Parse()

	if err := utils.RemoveEnvFile(); err != nil {
		log.Fatalf("error removing env file: %v", err)
	}

	discInfo, discID, err := fetchDiscInfoFromFlags()
	if err != nil {
		log.Fatalf("error parsing options: %v", err)
	}

	var gnuToc string
	if discID == "" {
		if gnuToc, discID, err = utils.GetTocAndDiscID(); err != nil {
			log.Fatalf("error retrieving disc ID: %v", err)
		}
	}
	cueFilePath := utils.CachePlaylistPath(discID)

	if utils.CheckIfPlaylistExists(cueFilePath) && !overwrite {
		return
	}

	if discInfo != nil && discID != "" {
		finalizeIfSuccess(discInfo, cueFilePath)
		return
	}
	var mbToc string
	if mbToc, err = utils.GetMusicBrainzDiscIDFromCmd(); err != nil {
		log.Fatalf("error retrieving MusicBrainz disc ID: %v", err)
	}

	if err = utils.CreateFolderIfNeeded(cueFilePath); err != nil {
		log.Fatalf("error creating folder for discID: %v", err)
	}

	// Fetch DiscInfo concurrently
	if discInfo, err = discinfo.FetchDiscInfoConcurrently(gnuToc, mbToc); err != nil {
		log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}

	finalizeIfSuccess(discInfo, cueFilePath)
}
