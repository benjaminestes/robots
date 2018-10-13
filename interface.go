// Package robots implements robots.txt file parsing and matching
// based on Google's specification. Read the full specification at:
// https://developers.google.com/search/reference/robots_txt.
//
// This specification assumes crawling a path is allowed by
// default. For both user agent and path matching, the longest match
// wins. For exampke, "/path" has precedence over "/". Path matches
// are case-sensitive; user agent matches are not. Acceptable
// metacharacters in paths specified by allow/disallow rules are "*" and
// "$". "*" matches anything, and "$" matches the end of a path. Metacharacters
// are not used in user agent matching. As a special case, a user agent
// pattern "*" matches any user agent.
//
// A generous parser is specified. Any line representing valid input
// is used, any invalid line is silently discarded. This is true even
// if the content parsed is in some unexpected format, like HTML.
//
// The rules for choosing which robots.txt file applies to a given URL
// are well-defined, and somewhat complex. Use Locate to get the URL
// of a robots.txt file that applies to a specific URL, and to any
// other URL whose crawling would be governed by that robots.txt file.
// It is the caller's responsibility to make sure you use the correct
// Robots object for a path; with Locate this should be easy.
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
func From(in io.Reader) (*Robots, error) {
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
