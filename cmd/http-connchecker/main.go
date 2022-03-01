package main

import (
	"io"
	"log"
	"net/http"
)

const (
	port string = ":8080"
)

func main() {
	http.HandleFunc("/hello", helloWorld)

	log.Println("Listening on port", port)
	err := http.ListenAndServe(port, nil)
	if err!= nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World!")
}