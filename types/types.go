package types

// DiscInfo contains metadata about a music disc (e.g., album or CD).
//
// Fields:
//   - ID (string): The unique identifier for the disc (e.g., MusicBrainz ID or custom disc ID).
//   - Artist (string): The name of the artist or band.
//   - Title (string): The title of the release (album name).
//   - ReleaseDate (string): The release date of the disc (e.g., "2024-01-01").
//   - Genre (string): The genre of the music (e.g., "Rock", "Pop").
//   - Tracks ([]string): A list of track titles in the release.
//   - CoverArtPath (string): The file path where the cover art image is stored (optional).
type DiscInfo struct {
	ID           string   // Unique ID for the disc
	Artist       string   // Artist or band name
	Title        string   // Title of the album or release
	ReleaseDate  string   // Release date of the disc
	Genre        string   // Genre of the album
	Tracks       []string // List of track titles
	CoverArtPath string   // Path to the cover art image
}

// MBRelease represents a MusicBrainz release. This struct is used for parsing MusicBrainz API responses.
//
// Fields:
//   - ID (string): The unique identifier of the release in MusicBrainz.
//   - Title (string): The title of the release (album name).
//   - Date (string): The release date in MusicBrainz format (e.g., "2024-01-01").
//   - ArtistCredit ([]struct{Name string}): A list of artist credits, with each artist's name as a string.
//   - Media ([]struct{ Tracks []struct{ Title string } }): Media tracks, with each track containing its title.
type MBRelease struct {
	ID           string `json:"id"`    // MusicBrainz release ID
	Title        string `json:"title"` // Release title
	Date         string `json:"date"`  // Release date in MusicBrainz format
	ArtistCredit []struct {
		Name string // Artist credit information
	} `json:"artist-credit"`
	Media []struct { // List of tracks in the release
		Tracks []struct{ Title string }
	} `json:"media"`
}

// ReleaseResult contains a list of releases returned by MusicBrainz in response to a query.
// Fields:
//   - Releases ([]MBRelease): A list of `MBRelease` structs that represent the releases matched by the query.
type ReleaseResult struct {
	Releases []MBRelease `json:"releases"` // List of matching releases
}
