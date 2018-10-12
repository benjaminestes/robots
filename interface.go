// Package robots implements robots.txt file parsing and matching
// based on Google's specification. Read the full specification at:
// https://developers.google.com/search/reference/robots_txt.
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
func From(in io.Reader) (Robots, error) {
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	return parse(string(buf)), nil
}

// Locate takes a string representing an absolute URL and returns the
// absolute URL of the robots.txt that would govern its crawlability
// (assuming such a file exists).
//
// Locate covers all special cases of the specification, including
// punycode domains, domain and protocol case-insensitivity, and
// default ports for certain protocols.
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
