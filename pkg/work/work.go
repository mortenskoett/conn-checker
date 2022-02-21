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

type UrlJob struct {
	Id  string
	Url string
}

func NewErrorCsvOutput(job UrlJob, err error) []string {
	return []string{job.Id, job.Url, err.Error()}
}

func NewErrorColumnNames() []string {
	return []string{"Id", "Original URL", "Error"}
}

func NewSuccessCsvOutput(job UrlJob, cr *conn.ConnectionResult) []string {
	return []string{job.Id, job.Url, cr.ReqUrl.String(), cr.EndUrl.String(), cr.Status, fmt.Sprint(cr.Redirects)}
}

func NewSuccessColumnNames() []string {
	return []string{"Id", "Original URL", "Request URL", "End URL", "Status", "Redirects"}
}

type Column uint16

const (
	// Column in the document
	IdCol  Column = 0
	UrlCol Column = 29
)

// Creates an empty channel that can receive UrlJobs and sets workerCount workers to take jobs from
// the queue.
func PrepareJobQueue(workerCount int, wg *sync.WaitGroup, tmpOutputDir, robotsOutputDir string) chan UrlJob {
	ch := make(chan UrlJob)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go urlWorker(ch, wg, tmpOutputDir, robotsOutputDir)
	}

	return ch
}

// The worker tries to parse the URL. If the operation succeeds then the worker attempts to connect
// to the url while collecting redirects. Both when failing or succeeding the worker writes the
// result to a separate file for each job.
func urlWorker(ch <-chan UrlJob, wg *sync.WaitGroup, tmpOutputDir, robotsOutputDir string) {
	defer wg.Done()
	var localpath string

	for job := range ch {
		parsedUrl, err := conn.ParseToUrl(job.Url)
		if err != nil {
			localpath = tmpOutputDir + job.Id + ".err"
			persist.PersistCsvLine(localpath, NewErrorCsvOutput(job, err))
			log.Print(job.Id, "error added: parsing to url:", parsedUrl, err)
			continue
		}

		result, err := conn.Connect(parsedUrl)
		if err != nil {
			localpath = tmpOutputDir + job.Id + ".err"
			persist.PersistCsvLine(localpath, NewErrorCsvOutput(job, err))
			log.Println(job.Id, "error added: connecting to site:", result, err)
			continue
		}

		// Download robots.txt
		if result.StatusCode == 200 {
			robotsTxtUrl := result.EndUrl.Scheme + "://" + result.EndUrl.Host + "/robots.txt"
			localpath = robotsOutputDir + job.Id + ".rob"
			err = conn.DownloadFileTo(robotsTxtUrl, localpath)
			if err != nil {
				log.Println("error reading robots.txt: ", err)
			}
		}

		// No error happened
		localpath = tmpOutputDir + job.Id + ".suc"
		persist.PersistCsvLine(localpath, NewSuccessCsvOutput(job, result))
		log.Println(job.Id, "success added: connection result:", result)
	}
}

// Fills work queue with UrlJobs that can be processed in parallel.
func ReadCsvIntoQueue(filepath string, ch chan<- UrlJob) error {
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
		ch <- UrlJob{Id: idEntry, Url: urlEntry}
	}
	return nil
}
