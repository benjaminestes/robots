package robots_test

import (
	"net/http"

	"github.com/benjaminestes/robots"
)

func ExampleRobots() {
	robotsURL, err := robots.Locate("https://www.example.com/page.html")
	if err != nil {
		// Handle error - couldn't parse input URL.
	}

	resp, err := http.Get(robotsURL)
	if err != nil {
		// Handle error.
	}
	defer resp.Body.Close()

	r, err := robots.From(200, resp.Body)
	if err != nil {
		// Handle error - couldn't read from input.
	}

	if r.Test("Googlebot", "/") {
		// You're good to crawl "/".
	}
	if r.Tester("Gooblebot")("/page.html") {
		// You're good to crawl "/page.html".
	}

	for _, sitemap := range r.Sitemaps() {
		// As the caller, we are responsible for ensuring that
		// the sitemap URL is in scope of the robots.txt file
		// we used before we try to access it.
		sitemapRobotsURL, err := robots.Locate(sitemap)
		if err != nil {
			// Couldn't parse sitemap URL - probably we should skip.
			continue
		}
		if sitemapRobotsURL == robotsURL && r.Test("Googlebot", sitemap) {
			resp, err := http.Get(sitemap)
			if err != nil {
				// Handle error.
			}
			defer resp.Body.Close()
			// ...do something with sitemap.
		}
	}
}
