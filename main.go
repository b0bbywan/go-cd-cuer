package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	mbToc, err := getMusicBrainzDiscID(gnuToc)
	log.Printf(mbToc)
	if err != nil {
		log.Fatalf("error retrieving disc ID: %v", err)
	}

	cueFilePath := cachePlaylistPath(discID)

	if _, err := os.Stat(cueFilePath); err == nil && !overwrite {
		saveEnvFile(cueFilePath)
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return
	}

	var discInfo *DiscInfo
	if err := os.MkdirAll(filepath.Dir(cueFilePath), os.ModePerm); err != nil {
		log.Fatalf("error creating folder for discID: %v", err)
	}
	// Fetch DiscInfo concurrently
	discInfo, err = fetchDiscInfoConcurrently(
		strings.Replace(gnuToc, " ", "+", -1),
		strings.Replace(mbToc, " ", "+", -1),
	)
	if err != nil {
		log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}

	finalizeIfSuccess(discInfo, cueFilePath)
}
