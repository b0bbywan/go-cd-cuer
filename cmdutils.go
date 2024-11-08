package main

import (
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

func getMusicBrainzDiscIDFromCmd() (string, error) {
	return runCommand("cd-discid", "--musicbrainz")
}

