package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
)

const (
	sites = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"
)

// Parse file
// Validate to some extend
// Check connection
// Write to separete files based on success
func main() {
	fmt.Println("Conn-checker started")

    f, err := os.Open(sites)
    if err != nil {
        log.Fatal(err)
    }

    defer f.Close()

    csvReader := csv.NewReader(f)
    for {
        line, err := csvReader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalln("an error occurred while parsing input file", err)
        }
        fmt.Printf("%+v\n", line)
    }

	url := "http://dr.dk"
	result, err := conn.Connect(url)
	if err != nil {
		log.Fatalln("error connecting to site", err)
	}

	fmt.Println(result)
}
