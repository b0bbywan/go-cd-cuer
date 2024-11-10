package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)
var (
	overwrite      bool

)

func init() {
	flag.BoolVar(&overwrite, "overwrite", false, "force regenerating the CUE file even if it exists")
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
