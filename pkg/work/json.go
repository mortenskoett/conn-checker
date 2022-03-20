package work

import (
	"log"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

type JsonUrlJob struct {
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
func PrepareJsonWorkQueue(workerCount uint8, wg *sync.WaitGroup) (chan JsonUrlJob, chan JobResultSuccess, chan JobResultError){
	jobCh := make(chan JsonUrlJob)
	successCh := make(chan JobResultSuccess)
	errorCh := make(chan JobResultError)

	// Start workers
	for i := uint8(0); i < workerCount; i++ {
		wg.Add(1)
		go jsonUrlWorker(jobCh, successCh, errorCh, wg)
	}

	return jobCh, successCh, errorCh
}

// The worker tries to parse the URL. If the operation succeeds then the worker attempts to connect
// to the url while collecting redirects. 
func jsonUrlWorker(jobChan <-chan JsonUrlJob, successChan chan<- JobResultSuccess, errChan chan<- JobResultError, wg *sync.WaitGroup) {
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

