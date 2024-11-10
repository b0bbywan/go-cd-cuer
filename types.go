package main

type DiscInfo struct {
	ID           string
	Artist       string
	Title        string
	ReleaseDate  string
	Genre        string
	Tracks       []string
	CoverArtPath string
}

type MBRelease struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Date           string `json:"date"`
	ArtistCredit   []struct{ Name string } `json:"artist-credit"`
	Media          []struct {
		Tracks []struct{ Title string }
	} `json:"media"`
}

type ReleaseResult struct {
    Releases []MBRelease `json:"releases"`
}

