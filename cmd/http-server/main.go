package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/work"
)

type ValidationResponse struct {
	Success []work.JobResultSuccess `json:"validations"`
	Errors []work.JobResultError `json:"errors"`
}

const (
	port string = ":8080"
	workerCount uint8 = 1
)

func main() {
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/validate", validate)

	log.Println("Listening on port", port)
	err := http.ListenAndServe(port, nil)
	if err!= nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	returnStatus(w, "pong", http.StatusOK)
}

func validate(w http.ResponseWriter, r *http.Request) {
	var urls []work.JsonUrlJob

	status, err := decodeJsonBodyInto(w, r, urls)
	if err != nil {
		returnStatus(w, err.Error(), status)
		return
	}

	log.Println(urls)

	// Create url job queue
	var wg sync.WaitGroup
	jobQueue, successOut, errorsOut := work.PrepareJsonWorkQueue(workerCount, &wg)

	// Add urls to queue
	for _, url := range urls {
		jobQueue <- url
	}

	// Wait for workers to finish processing urls
	close(jobQueue)
	wg.Wait()

	response := ValidationResponse{
		Success: make([]work.JobResultSuccess, 0, len(urls)/2), // TODO: Estimate
		Errors: make([]work.JobResultError, 0, len(urls)/2),
	}

	// Return collected results
	for success := range successOut {
		response.Success = append(response.Success, success)
	}
	for err := range errorsOut {
		response.Errors= append(response.Errors, err)
	}

	json, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

// Handling of input json request body
func decodeJsonBodyInto(w http.ResponseWriter, r *http.Request, outputVar interface{}) (int, error) {
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
			returnStatus(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType)
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

func returnStatus(w http.ResponseWriter, message string, status int) {
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

