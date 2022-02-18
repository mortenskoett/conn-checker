package conn

import (
	"fmt"
	"net/http"
)

const (
	maxRedirects = 10
)

type ConnectionResult struct {
	Status int
	ReqUrl string
	EndUrl string
	Redirects map[string]int // Url to statuscode
}

// Makes a GET request to the given URL. If URL is valid then a max of  n 
// redirects are followed to determine the status of the end url. 
func Connect(url string) (*ConnectionResult, error) {
	result := &ConnectionResult{
		Status: 0,
		ReqUrl: "",
		EndUrl: "",
		Redirects: make(map[string]int),
	}

    client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {

			// Collect data about redirects
			fromUrl := via[len(via)-1].URL.String()
			fromStatus := req.Response.StatusCode
			result.Redirects[fromUrl] = fromStatus // Set each redirect with status code
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Collect data after redirects
	result.Status = resp.StatusCode
	result.ReqUrl = url
	result.EndUrl = resp.Request.URL.String()
	result.Redirects[result.EndUrl] = resp.StatusCode // Set end url as final element in Redirects

	fmt.Println("CONNECTION RESULT: ", result)
	return result, nil
}