package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	envFile       = "/tmp/cd_var.env"
)

var (
	cacheLocation = fmt.Sprintf("%s/.cddb", os.Getenv("HOME"))
)

func generateCueFile(info *DiscInfo, cueFilePath string) error {
	file, err := os.Create(cueFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	content := fmt.Sprintf("REM DATE \"%s\"\nREM COVER \"%s\"\nPERFORMER \"%s\"\nTITLE \"%s\"\n",
		info.ReleaseDate, info.CoverArtPath, info.Artist, info.Title)

	for i, track := range info.Tracks {
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n",
			i+1, i+1, track)
	}

	_, err = file.WriteString(content)
	return err
}

func cachePlaylistPath(discID string) string {
	return filepath.Join(cacheLocation, discID, "playlist.cue")
}

// Cache cover art path
func cacheCoverArtPath(discID string) string {
	return filepath.Join(cacheLocation, discID, "cover.jpg")
}

func saveEnvFile(cueFile string) error {
	return os.WriteFile(envFile, []byte(fmt.Sprintf("CUE_FILE=%s", cueFile)), 0644)
}
