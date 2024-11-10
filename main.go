package main

import (
	"flag"
	"fmt"
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
func fetchDiscInfoFromFlags() (*DiscInfo, string, error) {
	// Enforce --musicbrainz with --disc-id
	if providedDiscID != "" && musicbrainzID == "" {
		return nil, "", fmt.Errorf("error: --disc-id option requires --musicbrainz to be set")
	}

	// If --musicbrainz is provided, fetch DiscInfo directly from MusicBrainz
	if musicbrainzID != "" {
		discInfo, err := fetchMusicBrainzReleaseByID(musicbrainzID)
		if err != nil {
			return nil, "", err
		}
		return discInfo, providedDiscID, nil
	}
	return nil, "", nil
}

func finalizeIfSuccess(discInfo *DiscInfo, cueFilePath string) {
	// Generate the CUE file and save
	if err := generateCueFile(discInfo, cueFilePath); err != nil {
		log.Fatalf("error: failed to generate CUE file: %v", err)
	}
	saveEnvFile(cueFilePath)
	log.Printf("info: Playlist generated at %s", cueFilePath)
}

func main() {
	flag.Parse()

	if err := removeEnvFile(envFile); err != nil {
		log.Fatalf("error removing env file: %v", err)
	}

	discInfo, discID, err := fetchDiscInfoFromFlags()
	if err != nil {
		log.Fatalf("error parsing options: %v", err)
	}

	var gnuToc string
	if discID == "" {
		if gnuToc, discID, err = getTocAndDiscID(); err != nil {
			log.Fatalf("error retrieving disc ID: %v", err)
		}
	}
	cueFilePath := cachePlaylistPath(discID)

	if checkIfPlaylistExists(cueFilePath) && !overwrite {
		return
	}

	if discInfo != nil && discID != "" {
		finalizeIfSuccess(discInfo, cueFilePath)
		return
	}
	var mbToc string
	if mbToc, err = getMusicBrainzDiscIDFromCmd(); err != nil {
		log.Fatalf("error retrieving MusicBrainz disc ID: %v", err)
	}

	if err = createFolderIfNeeded(cueFilePath); err != nil {
		log.Fatalf("error creating folder for discID: %v", err)
	}

	// Fetch DiscInfo concurrently
	if discInfo, err = fetchDiscInfoConcurrently(gnuToc, mbToc); err != nil {
		log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}

	finalizeIfSuccess(discInfo, cueFilePath)
}
