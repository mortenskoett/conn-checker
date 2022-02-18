package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

type Flattener interface {
	Flatten() []string
}

type OutputBase struct {
	Id string
	UnmodifiedInputUrl string
}

type ErrorOutputResult struct {
	OutputBase
	Error string
}

type SuccessOutputResult struct {
	OutputBase
	ConnectionResult *conn.ConnectionResult
}

func newErrorOutputResult(id, inputUrl, err string) Flattener {
	return &ErrorOutputResult{
		OutputBase: OutputBase{
			Id: id,
			UnmodifiedInputUrl: inputUrl,
		},
		Error: err,
	}
}

func newSuccessOutputResult(id, inputUrl string, connectionResult *conn.ConnectionResult) Flattener {
	return &SuccessOutputResult{
		OutputBase: OutputBase{
			Id: id,
			UnmodifiedInputUrl: inputUrl,
		},
		ConnectionResult: connectionResult,
	}
}

func (r *ErrorOutputResult) Flatten() []string {
	// Should match columnNames
	return []string {
		r.Id, r.UnmodifiedInputUrl, r.Error,
	}
}

func (r *SuccessOutputResult) Flatten() []string {
	// Should match columnNames
	return []string {
		r.Id,
		r.UnmodifiedInputUrl,
		r.ConnectionResult.ReqUrl,
		r.ConnectionResult.EndUrl,
		strconv.Itoa(r.ConnectionResult.Status),
		fmt.Sprint(r.ConnectionResult.Redirects),
	}
}

type Column uint16

const (
	// Column in the document
	IdCol Column = 0
	UrlCol Column = 29

	// Input file
	inputFilePath string = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"
	// inputFilePath string = "data/testdata.csv"

	// Output files
	outputSuccessPath string = "output/success.csv"
	outputErrorPath string = "output/errors.csv"
)

// Parse file
// Validate to some extend
// Check connection
// Write to separate files based on success
func main() {
	fmt.Println("Conn-checker started")

    f, err := os.Open(inputFilePath)
    if err != nil {
        log.Fatal("error while opening the file:", err)
    }
    defer f.Close()

    reader := csv.NewReader(f)

	// Read specification from first line
	specification, err := reader.Read()
	if err != nil {
		log.Fatalln("error while parsing first line of file:", err)
	}

	errorColumnNames := []string{specification[IdCol], specification[UrlCol], "Error"}
	successColumnNames := []string{
		specification[IdCol],
		specification[UrlCol],
		"Request Url",
		"End Url",
		"Status Code",
		"Redirects",
	}

	outputSuccessData := make([]Flattener, 0, 10000) // Arbitrary estimated size
	outputErrorData := make([]Flattener, 0, 5000) // Arbitrary estimated size

    for {
        line, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalln("error parsing input file entry:", line, err)
        }

		// Columns read from csv
		idEntry := line[IdCol]
		urlEntry := line[UrlCol]

		parsedUrl, err := parseUrl(urlEntry)
		if err != nil {
			log.Print("error added: parsing to url", parsedUrl)
			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
			continue
		}

		result, err := conn.Connect(parsedUrl.String())
		if err != nil {
			log.Println("error added: connecting:", result)
			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
			continue
		}

		// if result.Status >= 200 && result.Status < 300 {
			log.Println("connection result added:", result)
			outputSuccessData = append(outputSuccessData, newSuccessOutputResult(idEntry, urlEntry, result))
		// } else {
			// log.Println("error added: statuscode:", result)
			// outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
		// }
	}

	// Save to files when done collecting status codes
	persist(outputSuccessPath, outputSuccessData, successColumnNames)
	persist(outputErrorPath, outputErrorData, errorColumnNames)

	fmt.Println("Conn-checker done.")
}

// Attempts to parse the given string into a URL.
func parseUrl(u string) (*url.URL, error) {
	conformedUrl, err := conform(u)
	if err != nil {
		return nil, err
	}

	parsedUrl, err := url.Parse(conformedUrl)
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

// Persist data to specific location overwriting any file already there.
func persist(relPath string, fs []Flattener, columnNames []string) error {
	data := prepare(fs, columnNames)

    f, err := os.Create(relPath)
    if err != nil {
        log.Fatal("error while creating output file", err)
    }
    defer f.Close()

    writer := csv.NewWriter(f)

    defer writer.Flush()

    writer.WriteAll(data)
    if err != nil {
        log.Fatal("error while writing to output file", err)
    }

	log.Println("Persisted successfully to file: ", relPath)
	return nil
}

// Prepares the slice of Flatteners to be persisted
func prepare(fs []Flattener, columnNames []string) [][]string {
	data := make([][]string, 0, len(fs)+1) // +1 b/c of the first row of column names
	data = append(data, columnNames)

	for _, p := range fs {
		data = append(data, p.Flatten())
	}
	return data
}
