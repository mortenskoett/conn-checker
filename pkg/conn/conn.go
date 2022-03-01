package conn

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	maxRedirects       = 20
	connectTimeoutSecs = 10
)

type Redirect struct {
	Url	 	*url.URL
	Status 	int
}

type ConnectionResult struct {
	Status    string
	StatusCode int
	ReqUrl    *url.URL
	EndUrl    *url.URL
	Redirects []Redirect // Url to statuscode
}

// Makes a GET request to the given URL. If URL is valid then a max of
// 'maxRedirects' redirects are followed to determine the status of the end url.
func Connect(url *url.URL) (*ConnectionResult, error) {
	result := &ConnectionResult{
		Status:    "",
		StatusCode: 0,
		ReqUrl:    nil,
		EndUrl:    nil,
		Redirects: make([]Redirect, 0, 1),
	}

	client := &http.Client{
		Timeout: connectTimeoutSecs * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("max number of redirects reached")
			}

			// Collect data about redirects
			fromUrl := via[len(via)-1].URL
			fromStatus := req.Response.StatusCode
			result.Redirects = append(result.Redirects, Redirect{Url: fromUrl, Status: fromStatus}) // Set each redirect with status code
			return nil
		},
	}

	resp, err := client.Get(url.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Collect data after redirects
	result.Status = resp.Status
	result.StatusCode = resp.StatusCode
	result.ReqUrl = url
	result.EndUrl = resp.Request.URL
	result.Redirects = append(result.Redirects, Redirect{Url: result.EndUrl, Status: resp.StatusCode}) // Set end url as final element in Redirects

	return result, nil
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

// Attempts to parse the given string into URL format.
func ParseToUrl(u string) (*url.URL, error) {
	schemedUrl, err := addScheme(u)
	if err != nil {
		return nil, err
	}

	parsedUrl, err := url.Parse(schemedUrl)
	if err != nil {
		return nil, err
	}

	return parsedUrl, nil
}

// Makes asssumptions about input string and modifies it to be handlable as URL.
func addScheme(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("the given url was empty")
	}

	if !strings.HasPrefix(url, "http") {
		var sb strings.Builder
		sb.WriteString("http://")
		sb.WriteString(url)
		return sb.String(), nil
	}

	// Otherwise we assume nothing is wrong.
	return url, nil
}
