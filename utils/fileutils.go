package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	cacheLocation = fmt.Sprintf("%s/.cddb", os.Getenv("HOME"))
)

// checkIfPlaylistExists checks if the CUE file already exists.
func CheckIfPlaylistExists(cueFilePath string) bool {
	if _, err := os.Stat(cueFilePath); err == nil {
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return true
	}
	return false
}

// createFolderIfNeeded creates the necessary folder for the playlist file.
func CreateFolderIfNeeded(cueFilePath string) error {
	return os.MkdirAll(filepath.Dir(cueFilePath), os.ModePerm)
}

func CachePlaylistPath(discID string) string {
	return filepath.Join(cacheLocation, discID, "playlist.cue")
}

// Cache cover art path
func CacheCoverArtPath(discID string) string {
	return filepath.Join(cacheLocation, discID, "cover.jpg")
}
