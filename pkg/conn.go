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
	Redirects map[string]int // Url giving status code
}

func Connect(url string) (*ConnectionResult, error) {
	result := ConnectionResult{Redirects: make(map[string]int)}
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

		// fmt.Println("Output:", resp.Status)
		// fmt.Println("Req url", url)
		// fmt.Println("End url:", resp.Request.URL.String())

		// result := ConnectionResult{
		// 	Status: resp.Status,
		// 	ReqUrl: url,
		// 	EndUrl: resp.Request.URL.String(),
		// }

		result.Status = resp.Status
		result.ReqUrl = url
		result.EndUrl = resp.Request.URL.String()

		// Try going to next redirect
		if next := resp.Header.Get("Location"); next != "" {
			nextUrl = next
			result.Redirects[url] = resp.StatusCode
		} else {
			return &result, nil
		}
	}
	return &result, nil
}