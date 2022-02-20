package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/conn"
	"github.com/msk-siteimprove/conn-checker/pkg/persist"
	"github.com/msk-siteimprove/conn-checker/pkg/work"
)

const (
	// Input file
	// TODO
	// inputFileFile string = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"
	// inputFileFile string = "data/hometestdata_small.csv"
	// inputFileFile string = "data/hometestdata_very_small.csv"
	inputFileFile string = "data/hometestdata_big.csv"

	// Final output files
	outputSuccessFile string = "output/success.csv"
	outputErrorFile   string = "output/errors.csv"

	// Temporary output files
	tmpOutputDir     string = "output/tmp/"
	tmpSuccessSuffix string = ".suc"
	tmpErrorSuffix   string = ".err"

	// Number of goroutines
	workerCount int = 20
)

// Read in csv to job queue
// Workers process elements and each persist result to separate file
// Combine relevant results into errors, successes output files
func main() {
	fmt.Println("Conn-checker started")

	// Create dir to store temp files
	err := os.MkdirAll(tmpOutputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Create url job queue
	var wg sync.WaitGroup

	urlJobQueue := work.PrepareWorkQueue(workerCount, &wg, tmpOutputDir, outputSuccessFile, outputErrorFile)

	err = work.ReadCsvIntoQueue(inputFileFile, urlJobQueue)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for workers to finish processing urls
	close(urlJobQueue)
	wg.Wait()

	// Use Output format to write csv headers to output files
	connResultHeader := conn.ConnectionResult{
		ReqUrl:    "Request URL",
		EndUrl:    "End URL",
		Status:    "Status",
		Redirects: []conn.Redirect{conn.Redirect{Url: "Redirects", Status: -1}}}

	persist.PersistCsvLine(outputSuccessFile, work.NewSuccessCsvOutput(work.UrlJob{Id: "Id", Url: "Original URL"}, &connResultHeader))
	persist.PersistCsvLine(outputErrorFile, work.NewErrorCsvOutput(work.UrlJob{Id: "Id", Url: "Original URL"}, fmt.Errorf("Error")))

	// Combine tmp files together
	err = persist.Combine(tmpOutputDir, tmpSuccessSuffix, tmpErrorSuffix, outputSuccessFile, outputErrorFile)
	if err != nil {
		log.Println("error combining tmp files into output files:", err)
	}

	fmt.Println("Conn-checker finished")
}
