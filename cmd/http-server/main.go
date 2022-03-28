package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/utils"
	"github.com/msk-siteimprove/conn-checker/pkg/work"
)

type ValidationResponse struct {
	HttpSuccess []work.JobHttpSuccess`json:"http_success"`
	HttpErrors []work.JobHttpError `json:"http_errors"`
	OtherErrors[]work.JobOtherError `json:"other_errors"`
}

const (
	port string = ":8080"
	workerCount uint8 = 20
)

func main() {
	// End points
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/validate", validate)

	// Start server
	fmt.Println(utils.Logo())
	log.Println("Conn-checker starting")

	log.Println("Listening on port", port)
	err := http.ListenAndServe(port, nil)
	if err!= nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	returnJsonStatus(w, "pong", http.StatusOK)
}

// Validate json request and investigate HTTPS status codes and robotstxt of all URL's.
func validate(w http.ResponseWriter, r *http.Request) {
	var urls []work.JsonUrlJob

	status, err := decodeJsonBodyInto(w, r, &urls)
	if err != nil {
		returnJsonStatus(w, err.Error(), status)
		return
	}

	jobCount := len(urls)

	log.Println("Processing urls: ", jobCount)

	// Create url job queue
	var wg sync.WaitGroup
	jobQueue, httpSuccessOut, httpErrorsOut, otherErrorsOut := work.PrepareJsonWorkQueues(workerCount, &wg)

	response := ValidationResponse{
		HttpSuccess: make([]work.JobHttpSuccess, 0, jobCount/2), // TODO: Estimated sizes
		HttpErrors: make([]work.JobHttpError, 0, jobCount/4),
		OtherErrors: make([]work.JobOtherError, 0, jobCount/4),
	}

	// Setup single goroutine to synchronize output from workers into data structure.
	go func() {
		wg.Add(1)
		defer wg.Done()

		for i := 0 ; i < jobCount; i++ {
			select {
				case suc :=  <-httpSuccessOut:
					response.HttpSuccess = append(response.HttpSuccess, suc)
				case err :=  <- httpErrorsOut:
					response.HttpErrors = append(response.HttpErrors, err)
				case err :=  <- otherErrorsOut:
					response.OtherErrors = append(response.OtherErrors, err)
			}
		}
	}()

	// Add urls to queue to start processing
	for _, url := range urls {
		jobQueue <- url
	}

	// No more jobs to be added
	close(jobQueue) 

	// Wait for workers to finish processing urls
	wg.Wait()

	if err := returnJsonPayload(w, response); err == nil {
		log.Println("Conn-checker done. Json results successfully served to client:", jobCount)
	}
	return
}

// Handling and validation of input json request body
func decodeJsonBodyInto(w http.ResponseWriter, r *http.Request, outputVar interface{}) (int, error) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			returnJsonStatus(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType)
		}

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&outputVar)
	if err != nil {
		if errors.As(err, &unmarshalTypeError) {
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", 
								unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return  http.StatusBadRequest, errors.New(msg)

		} else if errors.As(err, &syntaxError) {
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return  http.StatusBadRequest, errors.New(msg)

		} else {
			msg := fmt.Sprintf("Bad request or internal server error: %s", err)
			return  http.StatusBadRequest, errors.New(msg)
		}
	}
	return -1, nil
}

func returnJsonPayload(w http.ResponseWriter, payload ValidationResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		returnJsonStatus(w, fmt.Sprintf("error happened during JSON encoding of response: %s", err), http.StatusInternalServerError)
		return err
	}
	return nil
}

func returnJsonStatus(w http.ResponseWriter, message string, status int) {
	log.Println(status, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := make(map[string]string)
	res["message"] = message

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatalf("error happened during JSON encoding of response: %s", err)
	}
}

