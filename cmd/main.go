package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

type CsvOutput = []string

func newErrorCsvOutput(job UrlJob, err error) CsvOutput {
	return []string{job.Id, job.Url, err.Error()}
}

func newSuccessCsvOutput(job UrlJob, cr *conn.ConnectionResult) CsvOutput {
	return []string{job.Id, job.Url, cr.ReqUrl, cr.EndUrl, cr.Status, fmt.Sprint(cr)}
}

type Column uint16

const (
	// Column in the document
	// IdCol  Column = 0
	// UrlCol Column = 29
	// TODO: For debugging
	IdCol  Column = 0
	UrlCol Column = 1

	// Input file
	// inputFileFile string = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"
	inputFileFile string = "data/hometestdata_small.csv"
	// inputFileFile string = "data/hometestdata_very_small.csv"

	// Output for temporary files
	tmpOutputDir string = "output/tmp/"

	// Final output files
	outputSuccessFile string = "output/success.csv"
	outputErrorFile   string = "output/errors.csv"

	// Number of goroutines
	workerCount int = 20
)

type UrlJob struct {
	Id  string
	Url string
}

// Read in csv to job queue
// Fill job queue
// Workers process elements and each persist result to separate file
// Combine relevant results into errors, successes output files
func main() {
	fmt.Println("Conn-checker started")

	// Create dir to store temp files
	err := os.MkdirAll(tmpOutputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	// Job queue
	urlJobsChan := make(chan UrlJob)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go urlWorker(urlJobsChan, &wg)
	}

	readCsvIntoQueue(inputFileFile, urlJobsChan)

	// Wait for workers to finish their work
	close(urlJobsChan)
	wg.Wait()

	// Stitch files together
	err = combine(tmpOutputDir, outputSuccessFile, outputErrorFile)
	if err != nil {
		log.Println("error combining tmp files into output files", err)
	}

	// TODO
	// Remove temporary files
	// err = os.RemoveAll(tmpOutputDir)
	// if err != nil {
	// 	log.Println(err)
	// }

	fmt.Println("Conn-checker finished")
}

// The worker tries to parse the url. If the operation succeeds then the worker attempts to connect
// to the url while collecting redirects. Both when failing or succeeding the worker writes the
// result to a separate file for each job.
func urlWorker(ch <-chan UrlJob, wg *sync.WaitGroup) {

	defer wg.Done()
	for job := range ch {
		parsedUrl, err := parseUrl(job.Url)
		if err != nil {
			log.Print(job.Id, "error added: parsing to url", parsedUrl, err)
			errorPath := tmpOutputDir + job.Id + ".err"
			persistSingle(errorPath, newErrorCsvOutput(job, err))
			continue
		}

		result, err := conn.Connect(parsedUrl.String())
		if err != nil {
			log.Println(job.Id, "error added: connecting to site:", result, err)
			errorPath := tmpOutputDir + job.Id + ".err"
			persistSingle(errorPath, newErrorCsvOutput(job, err))
			continue
		}

		// No error happened trying to get status code
		log.Println(job.Id, "success added: connection result:", result)
		successPath := tmpOutputDir + job.Id + ".suc"
		persistSingle(successPath, newSuccessCsvOutput(job, result))
	}
}

// Fills work queue with jobs and returns first line of csv containing column names
func readCsvIntoQueue(filepath string, ch chan<- UrlJob) {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal("error while opening the file:", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Read first line so it is not added to queue
	_, err = reader.Read()
	if err != nil {
		log.Fatalln("error while parsing first line of file:", err)
	}

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

		// Add job to queue
		ch <- UrlJob{Id: idEntry, Url: urlEntry}
	}
}

// Attempts to parse the given string into URL format.
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

// Reads all *.suc and *.err files from tmpFilesDir and combines them into files
// at successOutput and errorOutput.
func combine(tmpFilesDir, successOutput, errorOutput string) error {
	// success := make([][]strings, 0, 10000) // Arbitrary estimate
	// errors := make([][]strings, 0, 5000)   // Arbitrary estimate

	successes, err := os.Create(successOutput)
	if err != nil {
		return log.Fatal(err)
	}
	defer successes.Close()

	errors, err := os.Create(errorOutput)
	if err != nil {
		return log.Fatal(err)
	}
	defer errors.Close()

	tmpFiles, err := ioutil.ReadDir(tmpFilesDir)
	if err != nil {
		return log.Fatal(err)
	}

	for _, f := range tmpFiles {
		if strings.HasSuffix(f.Name, ".suc") {
		} else if strings.HasSuffix(f.Name, ".err") {
		}
	}

	return nil
}

// Persists a single row as CSV to the designated file which is overwritten.
func persistSingle(relPath string, data []string) error {
	return persistMultiple(relPath, [][]string{data})
}

// Persist data as CSV to specific location overwriting any file already there.
func persistMultiple(relPath string, data [][]string) error {

	f, err := os.Create(relPath)
	if err != nil {
		log.Fatal("error while creating output file", relPath, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)

	defer writer.Flush()

	writer.WriteAll(data)
	if err != nil {
		log.Println("error while writing to output file", err)
		return err
	}

	return nil
}

// Prepares the slice of Flatteners to be persisted
// func prepare(fs []Flattener, columnNames []string) [][]string {
// 	data := make([][]string, 0, len(fs)+1) // +1 b/c of the first row of column names
// 	data = append(data, columnNames)

// 	for _, p := range fs {
// 		data = append(data, p.Flatten())
// 	}
// 	return data
// }

// func run() {
// 	f, err := os.Open(inputFileFile)
// 	if err != nil {
// 		log.Fatal("error while opening the file:", err)
// 	}
// 	defer f.Close()

// 	reader := csv.NewReader(f)

// 	// Read specification from first line
// 	specification, err := reader.Read()
// 	if err != nil {
// 		log.Fatalln("error while parsing first line of file:", err)
// 	}

// 	errorColumnNames := []string{specification[IdCol], specification[UrlCol], "Error"}
// 	successColumnNames := []string{
// 		specification[IdCol],
// 		specification[UrlCol],
// 		"Request Url",
// 		"End Url",
// 		"Status Code",
// 		"Redirects",
// 	}

// 	outputSuccessData := make([]Flattener, 0, 10000) // Arbitrary estimated size
// 	outputErrorData := make([]Flattener, 0, 5000)    // Arbitrary estimated size

// 	for {
// 		line, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalln("error parsing input file entry:", line, err)
// 		}

// 		// Columns read from csv
// 		idEntry := line[IdCol]
// 		urlEntry := line[UrlCol]

// 		parsedUrl, err := parseUrl(urlEntry)
// 		if err != nil {
// 			log.Print("error added: parsing to url", parsedUrl)
// 			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
// 			continue
// 		}

// 		result, err := conn.Connect(parsedUrl.String())
// 		if err != nil {
// 			log.Println("error added: connecting:", result)
// 			outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
// 			continue
// 		}

// 		// if result.Status >= 200 && result.Status < 300 {
// 		log.Println("connection result added:", result)
// 		outputSuccessData = append(outputSuccessData, newSuccessOutputResult(idEntry, urlEntry, result))
// 		// } else {
// 		// log.Println("error added: statuscode:", result)
// 		// outputErrorData = append(outputErrorData, newErrorOutputResult(idEntry, urlEntry, err.Error()))
// 		// }
// 	}

// 	// Save to files when done collecting status codes
// 	persist(outputSuccessFile, outputSuccessData, successColumnNames)
// 	persist(outputErrorFile, outputErrorData, errorColumnNames)

// 	fmt.Println("Conn-checker done.")
// }
