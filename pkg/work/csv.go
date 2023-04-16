package work

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
	"github.com/msk-siteimprove/conn-checker/pkg/persist"
)

type CsvUrlJob struct {
	Id  string
	Url string
}

func NewErrorCsvOutput(job JsonUrlJob, err error) []string {
	return []string{job.Id, job.Url, err.Error()}
}

func NewErrorColumnNames() []string {
	return []string{"Id", "Original URL", "Error"}
}

func NewSuccessCsvOutput(job JsonUrlJob, cr *conn.ConnectionResult) []string {
	return []string{job.Id, job.Url, cr.ReqUrl.String(), cr.EndUrl.String(), cr.Status, fmt.Sprint(cr.Redirects)}
}

func NewSuccessColumnNames() []string {
	return []string{"Id", "Original URL", "Request URL", "End URL", "Status", "Redirects"}
}

type Column uint16

const (
	IdCol  Column = 0
	UrlCol Column = 1
)

// Creates an empty channel that can receive UrlJobs and sets workerCount workers to take jobs from
// the queue.
func PrepareCsvWorkQueue(workerCount int, wg *sync.WaitGroup, tmpOutputDir, successOutputPath, errorOutputPath string) chan JsonUrlJob {
	ch := make(chan JsonUrlJob)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go csvUrlWorker(ch, wg, tmpOutputDir, successOutputPath, errorOutputPath)
	}
	return ch
}

// The worker tries to parse the URL. If the operation succeeds then the worker attempts to connect
// to the url while collecting redirects. Both when failing or succeeding the worker writes the
// result to a separate file for each job.
func csvUrlWorker(ch <-chan JsonUrlJob, wg *sync.WaitGroup, tmpOutputDir, successOutputPath, errorOutputPath string) {
	defer wg.Done()
	for job := range ch {
		parsedUrl, err := conn.ParseToUrl(job.Url)
		if err != nil {
			log.Print(job.Id, "error added: parsing to url", parsedUrl, err)
			errorPath := tmpOutputDir + job.Id + ".err"
			persist.PersistCsvLine(errorPath, NewErrorCsvOutput(job, err))
			continue
		}

		result, err := conn.Connect(parsedUrl)
		if err != nil {
			log.Println(job.Id, "error added: connecting to site:", result, err)
			errorPath := tmpOutputDir + job.Id + ".err"
			persist.PersistCsvLine(errorPath, NewErrorCsvOutput(job, err))
			continue
		}

		// No error happened trying to get status code
		log.Println(job.Id, "success added: connection result:", result)
		successPath := tmpOutputDir + job.Id + ".suc"
		persist.PersistCsvLine(successPath, NewSuccessCsvOutput(job, result))
	}
}

// Fills work queue with UrlJobs that can be processed in parallel.
func ReadCsvIntoQueue(filepath string, ch chan<- JsonUrlJob) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error while opening the file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Read first line so it is not added to queue
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("error while parsing first line of file: %w", err)
	}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error parsing input file entry %s: %w", line, err)
		}

		// Columns read from csv
		idEntry := line[IdCol]
		urlEntry := line[UrlCol]

		// Add job to queue
		ch <- JsonUrlJob{Id: idEntry, Url: urlEntry}
	}
	return nil
}
