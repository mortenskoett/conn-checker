package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/text/encoding/charmap"
)

func main() {
	DownloadFileTo("https://www.sabes.it/robots.txt", "./output_mskk.txt")
}

// Download body of url directly to a local file which is created at 'filepath'.
func DownloadFileTo(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading file %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("could not download %s with status code %d", url, resp.StatusCode)
	}

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating local file %s: %w", filepath, err)
	}
	defer f.Close()

	// The Decoder transforms the source text to UTF-8 text. ISO8859-1 == Latin1.
	utf8Encoded := charmap.ISO8859_1.NewDecoder().Reader(resp.Body)

	_, err = io.Copy(f, utf8Encoded)
	return err
}