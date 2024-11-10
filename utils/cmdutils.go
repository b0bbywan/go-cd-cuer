package utils

import (
	"log"
	"os/exec"
	"strings"
)

// Execute a shell command and return its output as a string
func runCommand(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getDiscID() (string, error) {
	return runCommand("cd-discid")
}

func GetMusicBrainzDiscIDFromCmd() (string, error) {
	mbToc, err := runCommand("cd-discid", "--musicbrainz")
	if err != nil {
		return "", err
	}
	log.Printf("MB DiscID: %s", mbToc)
	return mbToc, nil

}

func GetTocAndDiscID() (string, string, error) {
	toc, err := getDiscID()
	if err != nil {
		return "", "", err
	}
	log.Printf("GNU DiscID: %s", toc)
	return toc, strings.Fields(toc)[0], nil
}