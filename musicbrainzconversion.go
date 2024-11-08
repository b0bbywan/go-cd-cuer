package main

import (
	"fmt"
	"strings"
	"strconv"
)

func parseDiscID(discID string) ([]int, int, error) {
	fields := strings.Fields(string(discID))
	if len(fields) < 3 {
		return nil, 0, fmt.Errorf("unexpected cd-discid output")
	}

	// Extract start sectors
	startSectors := make([]int, len(fields)-3)
	for i := 2; i < len(fields)-1; i++ {
		sector, err := strconv.Atoi(fields[i])
		if err != nil {
			return nil, 0, err
		}
		startSectors[i-2] = sector
	}

	// Extract last track length
	lastTrackLength, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return nil, 0, err
	}

	return startSectors, lastTrackLength, nil
}

func getMusicBrainzDiscID(discID string) (string, error) {
	startSectors, lastTrackLength, err := parseDiscID(discID)
	if err != nil {
		return "", err
	}
	numTracks := len(startSectors)
	totalSectors := startSectors[numTracks-1] + lastTrackLength

	var mbToc strings.Builder
	mbToc.WriteString(fmt.Sprintf("1 %d %d", numTracks, totalSectors))
	for _, sector := range startSectors {
		mbToc.WriteString(fmt.Sprintf(" %d", sector))
	}
	return mbToc.String(), nil
}
