package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/msk-siteimprove/conn-checker/pkg/utils"
	"github.com/msk-siteimprove/conn-checker/pkg/persist"
	"github.com/msk-siteimprove/conn-checker/pkg/work"
)

const (
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
	// Parse CLI flags
	inputFile := flag.String("file", "", "Path to the input .csv file.")
	flag.Parse()

	if *inputFile == "" {
		log.Println("Input file path is required.")
		printUsage()
		return
	}

	log.Println("Conn-checker started")

	log.Println("Creating temporary directories")
	err := os.MkdirAll(tmpOutputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(robotsOutputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Reading rows into job queue")
	var wg sync.WaitGroup
	urlJobQueue := work.PrepareCsvWorkQueue(workerCount, &wg, tmpOutputDir, outputSuccessFile, outputErrorFile)

	err = work.ReadCsvIntoQueue(*inputFile, urlJobQueue)
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

	log.Println("Conn-checker finished")
}

func printUsage() {
	fmt.Println(utils.Logo())
	fmt.Print("Description")
	fmt.Println(`
	Reads a bunch of URL end points from a csv file and returns another csv file containing the HTTP
	and robotstxt results from contacting each endpoint. The input should be formatted in rows of
	{id, url}. The URL does not need to be well formatted. Currently output is stored on disk
	incrementally as it is collected. Output is found in the ./output/ folder.`)

	fmt.Println("\nUsage")
	flag.PrintDefaults()
}

