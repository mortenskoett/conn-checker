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

const (
	port string = ":8080"
	workerCount = 1
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

func validate(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
			returnStatus(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType)
		}

	var urls []work.UrlJob

	err, status := decodeJsonBody(w, r, urls)
	if err != nil {
		returnStatus(w, err.Error(), status)
	}

	// Create url job queue
	var wg sync.WaitGroup
	urlJobCh, successCh, errorsCh := work.PrepareJobQueues(workerCount, &wg)

	// TODO Add data to work queue

	// Wait for workers to finish processing urls
	close(urlJobCh)
	wg.Wait()
	// TODO Return successes and errors as json
}

func ping(w http.ResponseWriter, r *http.Request) {
	returnStatus(w, "pong", http.StatusOK)
}

func decodeJsonBody(w http.ResponseWriter, r *http.Request, outputVar interface{}) (error, int) {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&outputVar)
	if err != nil {
		if errors.As(err, &unmarshalTypeError) {
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return errors.New(msg), http.StatusBadRequest

		} else if errors.As(err, &syntaxError) {
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return errors.New(msg), http.StatusBadRequest

		} else {
			msg := fmt.Sprintf("Bad request or internal server error", err) // TODO: Unsafe for public
			return errors.New(msg), http.StatusBadRequest
		}
	}
	return nil, 0
}

func returnStatus(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	res := make(map[string]string)
	res["message"] = message
	json, _ := json.Marshal(res)
	w.Write(json)
}