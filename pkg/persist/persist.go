package persist

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

// Reads all *.suc and *.err files from tmpFilesDir and combines them into separate files
// located at successOutput and errorOutput.
func Combine(tmpFilesDir, successSuffix, errorSuffix, successOutput, errorOutput string) error {
	// If the files do not exist, create them, and only append to the files
	successes, err := os.OpenFile(successOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer successes.Close()

	errors, err := os.OpenFile(errorOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer errors.Close()

	tmpFiles, err := os.ReadDir(tmpFilesDir)
	if err != nil {
		return err
	}

	// Write contents of each tmp file content to matching output file
	for _, f := range tmpFiles {
		if strings.HasSuffix(f.Name(), successSuffix) {
			tmpContents, err := os.ReadFile(tmpFilesDir + f.Name())
			if err != nil {
				return fmt.Errorf("error while reading tmp file: %w", err)
			}

			if _, err := successes.Write(tmpContents); err != nil {
				return fmt.Errorf("error while writing to output file: %w", err)
			}

		} else if strings.HasSuffix(f.Name(), errorSuffix) {
			tmpContents, err := os.ReadFile(tmpFilesDir + f.Name())
			if err != nil {
				return fmt.Errorf("error while reading tmp file: %w", err)
			}

			if _, err := errors.Write(tmpContents); err != nil {
				return fmt.Errorf("error while writing to output file: %w", err)
			}
		}
	}

	return nil
}

// Persists a single row as CSV to the designated file which is truncated in the process.
func PersistCsvLine(relPath string, data []string) error {
	return persistMultiple(relPath, [][]string{data})
}

// Persist data as CSV to specific location overwriting any file already there.
func persistMultiple(relPath string, data [][]string) error {

	f, err := os.Create(relPath)
	if err != nil {
		log.Fatal("error while creating output file", relPath, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)

	defer writer.Flush()

	writer.WriteAll(data)
	if err != nil {
		log.Println("error while writing to output file", err)
		return err
	}

	return nil
}

func RemoveFiles(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}
}
