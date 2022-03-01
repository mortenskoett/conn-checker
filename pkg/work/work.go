package work

import (
	"log"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

type UrlJob struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type JobResultSuccess struct {
	Id  string `json:"id"`
	EndUrl string `json:"end_url"`
	Status int `json:"status"`
}

type JobResultError struct {
	Id  string `json:"id"`
	ReqUrl string `json:"req_url"`
	EndUrl string `json:"end_url"`
	EndUrlStatus int `json:"end_url_status"`
	Suggestion string `json:"suggestion"`
}

// Creates an empty channel that can receive UrlJobs and sets workerCount workers to take jobs from
// the queue.
func PrepareJobQueues(workerCount int, wg *sync.WaitGroup) (chan UrlJob, JobResultSuccess, JobResultError){
	jobCh := make(chan UrlJob)
	successCh := make(chan JobResultSuccess)
	errorCh := make(chan JobResultError)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go urlWorker(jobCh, successCh, errorCh, wg)
	}

	return jobCh, <-successCh, <-errorCh
}

// The worker tries to parse the URL. If the operation succeeds then the worker attempts to connect
// to the url while collecting redirects. 
func urlWorker(jobChan <-chan UrlJob, successChan chan<- JobResultSuccess, errChan chan<- JobResultError, wg *sync.WaitGroup) {
	defer wg.Done()
	// var localpath string

	for job := range jobChan {
		parsedUrl, err := conn.ParseToUrl(job.Url)
		if err != nil {
			// TODO: Add to error channel
			log.Print(job.Id, "error added: parsing to url:", parsedUrl, err)
			continue
		}

		result, err := conn.Connect(parsedUrl)
		if err != nil {
			// TODO: Add to error channel
			log.Println(job.Id, "error added: connecting to site:", result, err)
			continue
		}

		// Download robots.txt
		if result.StatusCode == 200 {
			// TODO: Check robotstxt using smartjim/robots lib
			log.Println("error reading robots.txt: ", err)

			// robotsTxtUrl := result.EndUrl.Scheme + "://" + result.EndUrl.Host + "/robots.txt"
			// localpath = robotsOutputDir + job.Id + ".rob"
			// err = conn.DownloadFileTo(robotsTxtUrl, localpath)
			// if err != nil {
			// }
		}

		// TODO: Add to success chanel
		// No error happened
		log.Println(job.Id, "success added: connection result:", result)
	}
}

// Fills work queue with UrlJobs that can be processed in parallel.
// func ReadCsvIntoQueue(filepath string, ch chan<- UrlJob) error {
// 	f, err := os.Open(filepath)
// 	if err != nil {
// 		return fmt.Errorf("error while opening the file: %w", err)
// 	}
// 	defer f.Close()

// 	reader := csv.NewReader(f)

// 	// Read first line so it is not added to queue
// 	_, err = reader.Read()
// 	if err != nil {
// 		return fmt.Errorf("error while parsing first line of file: %w", err)
// 	}

// 	for {
// 		line, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return fmt.Errorf("error parsing input file entry %s: %w", line, err)
// 		}

// 		// Columns read from csv
// 		idEntry := line[IdCol]
// 		urlEntry := line[UrlCol]

// 		// Add job to queue
// 		ch <- UrlJob{Id: idEntry, Url: urlEntry}
// 	}
// 	return nil
// }
