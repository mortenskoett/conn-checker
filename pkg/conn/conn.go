package conn

import (
	"net/http"
)

const (
	maxRedirects = 10
)

type ConnectionResult struct {
	Status string
	ReqUrl string
	EndUrl string
	Redirects map[string]int // Url to statuscode
}

// Makes a GET request to the given URL if valid and follow n redirects
// to determine the status of the end url. 
func Connect(url string) (*ConnectionResult, error) {
	result := &ConnectionResult{Redirects: make(map[string]int)}
	nextUrl := url

    client := &http.Client{
      CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    } }
	
	for i := 0; i < maxRedirects; i++ {
		resp, err := client.Get(nextUrl)
		if err != nil {
			return nil, err
		}
		
		defer resp.Body.Close()

		result.Status = resp.Status
		result.ReqUrl = url
		result.EndUrl = resp.Request.URL.String()
		result.Redirects[result.EndUrl] = resp.StatusCode

		// Try going to next redirect
		if next := resp.Header.Get("Location"); next != "" {
			nextUrl = next
		} else {
			return result, nil
		}
	}
	return result, nil
}