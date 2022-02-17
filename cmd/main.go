package main

import (
	"fmt"
	"log"

	conn "github.com/msk-siteimprove/conn-checker/pkg"
)

func main() {
	fmt.Println("Conn-checker started")

	// Parse file

	// Validate to some extend

	// Check connection

	// Write to separete files

	// client := &http.Client{
	// 	// CheckRedirect: func(req *http.Request, via[]*http.Request) error {
		// 	return http.ErrUseLastResponse // Will return on first redirect or error
		// },
	// }


	url := "http://www.dr.dk"
	// resp, err := client.Get(url)
	// if err != nil {
	// 	log.Fatalln("error occured while connecting to site:", err)
	// }

	result, err := conn.Connect(url)
	if err != nil {
		log.Fatalln("error connecting to site", err)
	}

	fmt.Println(result)


	// fmt.Println(resp.Request.URL)
// 	fmt.Println("Output:", resp.Status)
// 	fmt.Println("Req url", url)
// 	fmt.Println("End url:", resp.Request.URL)
}
