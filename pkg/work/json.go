package work

import (
	"fmt"
	"log"
	"sync"

	"github.com/jimsmart/grobotstxt"
	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

var (
	// TODO: Should be given as arg to API
	agents = []string{"ConnCheckerBot/1.0"}
)

type JsonUrlJob struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type JobHttpSuccess struct {
	Id  string `json:"id"`
	EndUrl string `json:"end_url"`
	HttpStatus int `json:"http_status"`
	RobotsOk bool `json:"robots_ok"`
}

type JobHttpError struct {
	Id  string `json:"id"`
	ReqUrl string `json:"req_url"`
	Suggestion string `json:"suggestion"`
}

type JobOtherError struct {
	Id  string `json:"id"`
	Message string `json:"message"`
}

// Creates an empty channel that can receive UrlJobs and sets workerCount workers to take jobs from
// the queue.
func PrepareJsonWorkQueues(workerCount uint8, wg *sync.WaitGroup) (chan JsonUrlJob, chan JobHttpSuccess,
																	chan JobHttpError, chan JobOtherError){
	jobCh := make(chan JsonUrlJob)
	httpSuccessCh := make(chan JobHttpSuccess)
	httpErrorCh := make(chan JobHttpError)
	otherErrorCh := make(chan JobOtherError)

	// Start workers
	for i := uint8(0); i < workerCount; i++ {
		wg.Add(1)
		go jsonUrlWorker(jobCh, httpSuccessCh, httpErrorCh, otherErrorCh, wg)
	}

	return jobCh, httpSuccessCh, httpErrorCh, otherErrorCh
}

// The worker tries to parse the URL. If the operation succeeds then the worker attempts to connect
// to the url while collecting redirects. Robotstxt is also tested for crawlability with given agent.
func jsonUrlWorker(jobChan <-chan JsonUrlJob, httpSuccessChan chan<- JobHttpSuccess, 
					httpErrChan chan<- JobHttpError, otherErrorChan chan<- JobOtherError, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobChan {
		parsedUrl, err := conn.ParseToUrl(job.Url)
		if err != nil {
			msg := fmt.Sprint("error parsing to url:", err.Error())
			log.Println("Job:", job.Id, msg)
			otherErrorChan <- JobOtherError{Id: job.Id, Message: msg}
			continue
		}

		// Expects valid URL at this point

		result, err := conn.Connect(parsedUrl)
		if err != nil {
			log.Println(job.Id, "error connecting to site:", err)
			httpErrChan <- JobHttpError {
				Id: job.Id,
				ReqUrl: job.Url,
				Suggestion: err.Error(),
			}
			continue
		}

		// Investigate robots.txt
		robotsTxt := result.EndUrl.Scheme + "://" + result.EndUrl.Host + "/robots.txt"
		isAllowed := grobotstxt.AgentsAllowed(robotsTxt, agents, result.EndUrl.String())

		// Return https status results
		log.Println(job.Id, "OK http success result:", result.Status, result.EndUrl)
		httpSuccessChan <- JobHttpSuccess {
			Id: job.Id,
			EndUrl: result.EndUrl.String(),
			HttpStatus: result.StatusCode,
			RobotsOk : isAllowed,
		}
	}
}

