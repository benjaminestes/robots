// Package robots implements robots.txt parsing and matching based on
// Google's specification. For a robots.txt primer, please read the
// full specification at:
// https://developers.google.com/search/reference/robots_txt.
//
// What clients need to think about
//
// Clients of this package have one obligation: when testing whether a
// URL can be crawled, use the correct robots.txt file. The
// specification uses scheme, port, and punycode variations to define
// which URLs are in scope.
//
// To get the right robots.txt file, use Locate. Locate takes as its
// only argument the URL you want to access. It returns the URL of the
// robots.txt file that governs access. Locate will always return a
// single unique robots.txt URL for all input URLs sharing a scope.
//
// In practice, a client pattern for testing whether a URL is
// accessible would be: a) Locate the robots.txt file for the URL; b)
// check whether you have fetched data for that robots.txt file; c) if
// yes, use the data to Test the URL against your user agent; d) if
// no, fetch the robots.txt data and try again.
//
// For details, see "File location & range of validity" in the
// specification:
// https://developers.google.com/search/reference/robots_txt#file-location--range-of-validity.
//
// How bad input is handled
//
// A generous parser is specified. A valid line is accepted, and an
// invalid line is silently discarded. This is true even if the
// content parsed is in an unexpected format, like HTML.
//
// For details, see "File format" in the specification:
// https://developers.google.com/search/reference/robots_txt#file-format
package robots

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

// From produces a Robots object from a robots.txt file represented as
// an io.Reader.
//
// The attitude of the specification is permissive concerning parser
// errors: all valid input is accepted, and invalid input is silently
// rejected without failing. Therefore, From will only signal an error
// condition if it fails to read from the input at all.
func From(status int, in io.Reader) (*Robots, error) {
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	data := parse(string(buf))
	return makeRobots(status, data), nil
}

// Locate takes a string representing an absolute URL and returns the
// absolute URL of the robots.txt that would govern its crawlability
// (assuming such a file exists).
//
// Locate covers all special cases of the specification, including
// punycode domains, domain and protocol case-insensitivity, and
// default ports for certain protocols. It is guaranteed to produce
// the same robots.txt URL for any input URLs that share a scope.
func Locate(rawurl string) (string, error) {
	const (
		httpPort  = ":80"
		httpsPort = ":443"
		ftpPort   = ":21"
	)

	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if !u.IsAbs() {
		return "", fmt.Errorf("expected absolute URL, got: %s", rawurl)
	}

	switch {
	// do these need to be case-insensitive?
	case u.Scheme == "http" && strings.HasSuffix(u.Host, httpPort):
		u.Host = u.Host[:len(u.Host)-len(httpPort)]
	case u.Scheme == "https" && strings.HasSuffix(u.Host, httpsPort):
		u.Host = u.Host[:len(u.Host)-len(httpsPort)]
	case u.Scheme == "ftp" && strings.HasSuffix(u.Host, ftpPort):
		u.Host = u.Host[:len(u.Host)-len(ftpPort)]
	default:
		// Otherwise, the port stays put. Non-default ports
		// require their own robots.txt file.
	}
	// FIXME: Deal with error
	u.Host, _ = idna.ToUnicode(u.Host)
	u.Host = strings.ToLower(u.Host)
	u.Scheme = strings.ToLower(u.Scheme)
	return u.Scheme + "://" + u.Host + "/robots.txt", nil
}
