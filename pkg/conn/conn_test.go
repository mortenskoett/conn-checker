package conn

import (
	"fmt"
	"log"
	"testing"
)

func TestConnect(t *testing.T) {
	urlstr := "https://asfo.sanita.fvg.it"

	url, err  := ParseToUrl(urlstr)
	if err != nil {
		log.Fatal(err)
	}

	res, err := Connect(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}