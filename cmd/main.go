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

type Persister interface {
	FormatForCsv() []string
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

func (r *ErrorOutputResult) FormatForCsv() []string {
	return []string{}
}

func (r *SuccessOutputResult) FormatForCsv() []string {
	return []string{}
}

func newErrorOutputResult(id, inputUrl, err string) Persister {
	return &ErrorOutputResult{
		OutputBase: OutputBase{
			Id: id,
			UnmodifiedInputUrl: inputUrl,
		},
		Error: err,
	}
}

func newSuccessOutputResult(id, inputUrl string, connectionResult *conn.ConnectionResult) Persister {
	return &SuccessOutputResult{
		OutputBase: OutputBase{
			Id: id,
			UnmodifiedInputUrl: inputUrl,
		},
		ConnectionResult: connectionResult,
	}
}

type Column uint16

const (
	// Column in the document
	IdCol Column = 0
	UrlCol Column = 29

	// Input file
	urlListInputPath string = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"

	// Output files
	outputSuccessPath string = "output/success.csv"
	outputErrorPath string = "output/errors.csv"
	// outputFailedPath = "output/failed.csv"
)

var (
	outputSuccessData = make([]Persister, 10000) // Arbitrary estimated sizes
	outputErrorData = make([]Persister, 10000) // Arbitrary estimated sizes
	// outputFailedData = make([]*conn.ConnectionResult, 5000)
)

// Parse file
// Validate to some extend
// Check connection
// Write to separate files based on success
func main() {
	fmt.Println("Conn-checker started")

    f, err := os.Open(urlListInputPath)
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

		// Columns read from csv
		idEntry := line[IdCol]
		urlEntry := line[UrlCol]

		parsedUrl, err := parseUrl(urlEntry)
		if err != nil {
			fmt.Print("Appended parsedUrl error", parsedUrl)
			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
			log.Println("error while parsing to URL format:", err)
			continue
		}

		result, err := conn.Connect(parsedUrl.String())
		if err != nil {
			fmt.Print("Appended connect error", result)
			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
			log.Println("error while connecting to site:", parsedUrl.String(), err)
			continue
		}

		if result.Status == 200 {
			fmt.Println("Appended success", result)
			outputSuccessData = append(outputSuccessData, newSuccessOutputResult(idEntry, urlEntry, result))
		} else {
			fmt.Println("Appended error", result)
			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
		}
	}
		// persist(outputSuccessData, outputSuccessPath)
		persist(outputErrorPath, outputErrorData)
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

func prepare(persister []Persister, dataSize int) [][]string {
	data := make([][]string, dataSize)
	for _, p := range persister {
		data = append(data, p.FormatForCsv())
	}
	return data
}

func persist(relPath string, ps []Persister) error {
	data := [][]string{
			{"vegetables", "fruits"},
			{"carrot", "banana"},
			{"potato", "strawberry"},
		}

    // create a file
    file, err := os.Create(relPath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // initialize csv writer
    writer := csv.NewWriter(file)

    defer writer.Flush()

    // write all rows at once
    writer.WriteAll(data)

    // write single row
    extraData := []string{"lettuce", "raspberry"}
    writer.Write(extraData)

	return nil
}