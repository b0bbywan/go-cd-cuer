package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: failed to fetch from URL %s, status code: %d", url, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func makeGnuRequest(client *http.Client, url, userAgent string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	return client.Do(req)
}
