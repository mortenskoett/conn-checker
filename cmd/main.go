package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

type Column uint16

const (
	// Column in the document
	IdCol Column = 0
	UrlCol Column = 29

	// Input file
	sites = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"

	// Output files
	outputSuccessPath = "output/success.csv"
	outputFailedPath = "output/failed.csv"
)

var (
	outputSuccessData = make([]*conn.ConnectionResult, 10000) // Arbitrary estimated sizes
	outputFailedData = make([]*conn.ConnectionResult, 5000)
)

// Parse file
// Validate to some extend
// Check connection
// Write to separate files based on success
func main() {
	fmt.Println("Conn-checker started")

    f, err := os.Open(sites)
    if err != nil {
        log.Fatal("error while opening the file:", err)
    }

    defer f.Close()

    reader := csv.NewReader(f)
    for {
        line, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalln("error while parsing input file entry:", err)
        }

		parsedUrl, err := parseUrl(line[UrlCol])
		if err != nil {
			log.Println("error while parsing to URL format:", err)
			continue
		}

		// fmt.Println(parsedUrl.String())
		// fmt.Println("before:", conformedUrl)
		// fmt.Println("after:", parsedUrl.Host)

		result, err := conn.Connect(parsedUrl.String())
		if err != nil {
			log.Println("error while connecting to site:", parsedUrl.String(), err)
			continue
		}

		if result.Status != 200 {
			fmt.Println(result.Status)
		}
		// outputSuccessData = append(outputSuccessData, result)
	}

		// persist(outputSuccessData, outputSuccessPath)
}

// Attempts to parse the given string into a URL.
func parseUrl(u string) (*url.URL, error) {
	conformedUrl, err := conform(u)
	if err != nil {
		return nil, err
	}

	parsedUrl, err := url.ParseRequestURI(conformedUrl)
	if err != nil {
		return nil, err
	}

	return parsedUrl, err
}

// Makes asssumptions about input string and modifies it to be handlable as URL.
func conform(url string) (string, error) {
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

func persist(data []*conn.ConnectionResult, relPath string) error {
	return nil
}