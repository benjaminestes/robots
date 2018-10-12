package robots

import (
	"fmt"
	"net/http"

	"github.com/benjaminestes/robots"
)

func ExampleRobots() {
	robotsURL := robots.Locate("https://www.example.com/page.html")

	resp, err := http.Get(robotsURL)
	if err != nil {
		// Handle error.
	}
	defer resp.Body.Close()

	r := robots.From(resp.Body)
	if r.Test("Googlebot", "/") {
		// You're good to crawl.
	}
	if r.Tester("Gooblebot")("/page.html") {
		// You're good to crawl.
	}

	for _, sitemap := range r.Sitemaps {
		// Check that sitemap URL is in scope of the robots.txt file
		// we used.
		if r.Locate(sitemap) == robotsURL && r.Test("Googlebot", sitemap) {
			resp, err := http.Get(sitemap)
			// ...check errors, do something with sitemap, close resp.
		}
	}
}
