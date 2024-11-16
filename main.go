package main

import (
	"flag"
	"log"

	"github.com/b0bbywan/go-disc-cuer/cue"
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

func main() {
	flag.Parse()

	if _, err := cue.GenerateWithOptions(musicbrainzID, providedDiscID, overwrite); err != nil {
		log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}
}
