package utils

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

// checkIfPlaylistExists checks if the CUE file already exists.
func CheckIfPlaylistExists(cueFilePath string) bool {
	if _, err := os.Stat(cueFilePath); err == nil {
		SaveEnvFile(cueFilePath)
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return true
	}
	return false
}

// removeEnvFile removes the environment file if it exists.
func RemoveEnvFile() error {
	if err := os.Remove(envFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing env file: %v", err)
	}
	return nil
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

func SaveEnvFile(cueFile string) error {
	return os.WriteFile(envFile, []byte(fmt.Sprintf("CUE_FILE=%s", cueFile)), 0644)
}
