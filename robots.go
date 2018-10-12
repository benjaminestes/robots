package robots

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

type robots []*agent

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

func (r robots) Test(a, p string) bool {
	return r.Tester(a)(p)
}

func (r robots) Tester(a string) func(string) bool {
	agent, ok := r.bestAgent(a)
	if !ok {
		return func(_ string) bool {
			return true
		}
	}
	return func(path string) bool {
		for _, member := range agent.group {
			if member.match(path) {
				return member.allow
			}
		}
		return true
	}
}

func (r robots) bestAgent(a string) (*agent, bool) {
	for _, agent := range r {
		if agent.match(a) {
			return agent, true
		}
	}
	return nil, false
}
