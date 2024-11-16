package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/b0bbywan/go-disc-cuer/config"
)

var (
	cacheLocation = config.CacheLocation
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
