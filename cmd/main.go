package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/persist"
	"github.com/msk-siteimprove/conn-checker/pkg/work"
)

const (
	// Input file
	// inputFileFile string = "data/d09adf99-dc10-4349-8c53-27b1e5aa97b6.csv"
	inputFileFile string = "data/testdata.csv"

	// Final output files
	outputSuccessFile string = "output/success.csv"
	outputErrorFile   string = "output/errors.csv"

	// Temporary output files
	tmpOutputDir     string = "output/tmp/"
	robotsOutputDir	 string = "output/robots/"
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
	err = os.MkdirAll(robotsOutputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Create url job queue
	var wg sync.WaitGroup
	urlJobQueue := work.PrepareJobQueue(workerCount, &wg, tmpOutputDir, robotsOutputDir)
	err = work.ReadCsvIntoQueue(inputFileFile, urlJobQueue)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for workers to finish processing urls
	close(urlJobQueue)
	wg.Wait()

	// Write out first row as column names of output files
	persist.PersistCsvLine(outputSuccessFile, work.NewSuccessColumnNames())
	persist.PersistCsvLine(outputErrorFile, work.NewErrorColumnNames())

	// Combine tmp files together
	err = persist.Combine(tmpOutputDir, tmpSuccessSuffix, tmpErrorSuffix, outputSuccessFile, outputErrorFile)
	if err != nil {
		log.Fatal("error combining tmp files into output files:", err)
	}

	fmt.Println("Conn-checker finished")
}
