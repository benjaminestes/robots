package robots

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

func From(in io.Reader) (robots, error) {
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	return parse(string(buf)), nil
}

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
