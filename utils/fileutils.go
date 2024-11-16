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

// CheckIfPlaylistExists checks if a CUE playlist file already exists at the specified path.
//
// Parameters:
//   - cueFilePath (string): The path to the CUE file that needs to be checked for existence.
//
// Returns:
//   - bool: True if the playlist file exists, false otherwise.
//   - The function also logs an informational message if the file exists.
func CheckIfPlaylistExists(cueFilePath string) bool {
	if _, err := os.Stat(cueFilePath); err == nil {
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return true
	}
	return false
}

// CreateFolderIfNeeded ensures that the folder for the playlist file exists. If the folder does not exist, it creates it.
//
// Parameters:
//   - cueFilePath (string): The path to the CUE file whose folder needs to be created if it doesn't already exist.
//
// Returns:
//   - error: Any error encountered during the folder creation process.
func CreateFolderIfNeeded(cueFilePath string) error {
	return os.MkdirAll(filepath.Dir(cueFilePath), os.ModePerm)
}

// CachePlaylistPath generates the file path where the playlist CUE file is cached based on the disc ID.
//
// Parameters:
//   - discID (string): The disc ID used to generate the path for the cached playlist CUE file.
//
// Returns:
//   - string: The generated file path for the cached CUE file.
func CachePlaylistPath(discID string) string {
	return filepath.Join(cacheLocation, discID, "playlist.cue")
}

// CacheCoverArtPath generates the file path where the cover art image is cached based on the disc ID.
//
// Parameters:
//   - discID (string): The disc ID used to generate the path for the cached cover art file.
//
// Returns:
//   - string: The generated file path for the cached cover art image.
func CacheCoverArtPath(discID string) string {
	return filepath.Join(cacheLocation, discID, "cover.jpg")
}
