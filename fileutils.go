package main

import (
	"fmt"
	"log"
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

	if info.CoverArtPath == "" {
		coverFilePath := cacheCoverArtPath(filepath.Dir(cueFilePath))
		if fetchCoverArt(info.ID, coverFilePath) == nil {
			info.CoverArtPath = coverFilePath
		}
	}

	content := fmt.Sprintf("REM DATE \"%s\"\nREM COVER \"%s\"\nPERFORMER \"%s\"\nTITLE \"%s\"\n",
		info.ReleaseDate, info.CoverArtPath, info.Artist, info.Title)

	for i, track := range info.Tracks {
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n",
			i+1, i+1, track)
	}

	_, err = file.WriteString(content)
	return err
}

// checkIfPlaylistExists checks if the CUE file already exists.
func checkIfPlaylistExists(cueFilePath string) bool {
	if _, err := os.Stat(cueFilePath); err == nil {
		saveEnvFile(cueFilePath)
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return true
	}
	return false
}

// removeEnvFile removes the environment file if it exists.
func removeEnvFile(envFile string) error {
	if err := os.Remove(envFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing env file: %v", err)
	}
	return nil
}

// createFolderIfNeeded creates the necessary folder for the playlist file.
func createFolderIfNeeded(cueFilePath string) error {
	return os.MkdirAll(filepath.Dir(cueFilePath), os.ModePerm)
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
