package utils

import (
	"fmt"
	"log"
	"strings"

	"go.uploadedlobster.com/discid"
)

func GetTocAndDiscID(disc discid.Disc) (string, string, error) {
	gnuToc, err := tocToGnu(disc)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate GNU TOC: %w", err)
	}

	discID := disc.FreedbID()
	log.Printf("GNU TOC: %s", gnuToc)
	return gnuToc, discID, nil
}

// GetMusicBrainzDiscIDFromCmd retrieves the TOC string for MusicBrainz
func GetMusicBrainzTOC(disc discid.Disc) (string, error) {
	mbToc := disc.TOCString()
	log.Printf("MusicBrainz TOC: %s", mbToc)
	return mbToc, nil
}

func tocToGnu(disc discid.Disc) (string, error) {
	// Get FreeDB ID
	freedbID := disc.FreedbID()
	// Get the number of tracks
	trackCount := disc.LastTrackNumber()

	// Collect track offsets
	offsets := []string{freedbID, fmt.Sprintf("%d", trackCount)}
	for i := 1; i <= trackCount; i++ {
		track, err := disc.Track(i)
		if err != nil {
			return "", err
		}
		offsets = append(offsets, fmt.Sprintf("%d", track.Offset))
	}

	// Append the disc duration as an integer
	offsets = append(offsets, fmt.Sprintf("%d", int(disc.Duration().Seconds())))

	// Join the components with spaces
	return strings.Join(offsets, " "), nil
}
