package conn

import (
	"fmt"
	"log"
	"testing"
)

func TestConnect(t *testing.T) {
	url := "https://asfo.sanita.fvg.it"
	res, err := Connect(url)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}