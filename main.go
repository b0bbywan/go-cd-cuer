package main

import (
	"flag"
	"github.com/b0bbywan/go-cd-cuer/cue"
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

func main() {
	flag.Parse()

	if _, err := cue.Generate(musicbrainzID, providedDiscID, overwrite); err != nil {
		log.Fatalf("error: failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}
}
