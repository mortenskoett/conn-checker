package main

import (
	"fmt"

	"github.com/jimsmart/grobotstxt"
)

func main() {
// Contents of robots.txt file.
robotsTxt := `
    # robots.txt with restricted area

    User-agent: FooBot/1.0
    Disallow: /members/*

    User-agent: SiteimproveBot
    Disallow: /whoknows/*

    Sitemap: http://example.net/sitemap.xml
`

// Target URI.
uri := "http://example.net/members/index.html"

agents := []string{"FooBot/1.0", "SiteimproveBot"}

// Is bot allowed to visit this page?
// ok := grobotstxt.AgentAllowed(robotsTxt, "FooBot/1.0", uri)
ok1 := grobotstxt.AgentsAllowed(robotsTxt, agents, uri)

// fmt.Println("result ok: ", ok)
fmt.Println("result ok1: ", ok1)
}