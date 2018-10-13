package robots

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// member represents a group-member record as defined in Google's
// specification. After its path has been set, the compile() method
// must be called prior to use.
type member struct {
	allow   bool
	path    string
	pattern *regexp.Regexp
}

// Check whether the given path is matched by this record.
func (m *member) match(path string) bool {
	return m.pattern.MatchString(path)
}

// A group-member record specifies a path to which it
// applies. Internally to this package, we need an efficient way of
// matching that path, which possibly includes metacharacters * and
// $. compile() compiles the given path to a regular expression
// denoting the patterns we will accept.
func (m *member) compile() {
	// This approach to handling matches is derived from temoto's:
	// https://github.com/temoto/robotstxt/blob/master/parser.go
	pattern := regexp.QuoteMeta(m.path)
	pattern = "^" + pattern // But with an added start-of-line
	pattern = strings.Replace(pattern, `\*`, `.*`, -1)
	pattern = strings.Replace(pattern, `\$`, `$`, -1)
	// FIXME: What do I do in case of error?
	r, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("oh no! %v\n", err)
	}
	m.pattern = r
}

// A group is an ordered list of members. The members are ordered from
// longest path to shortest path. This allows efficient matching of
// paths to members: when evaluated sequentially, the first match must
// be the longest.
type group struct {
	members []*member
}

func (g *group) addMember(m *member) {
	// Maintain type invariant: a member must have its pattern
	// compiled before use.
	m.compile()
	g.members = append(g.members, m)
	// Maintain type invariant: the members of a group must always
	// be sorted by length of path, descending.
	sort.Slice(g.members, func(i, j int) bool {
		return len(g.members[i].path) > len(g.members[j].path)
	})
}

// An agent represents a group of rules that a named robots agent
// might match. Its compile() method must be called prior to use.
type agent struct {
	name    string
	group   group
	pattern *regexp.Regexp
}

// Test whether the given robots agent string matches this agent.
func (a *agent) match(name string) bool {
	return a.pattern.MatchString(strings.ToLower(name))
}

// A agent specifies a robots agent to which it applies. Internally to
// this package, we need an efficient way of matching that name, which
// includes no metacharacters. However, we will treat the special case
// "*" as matching all agents for which no other match
// exists. compile() compiles the amended name to a regular expression
// denoting the patterns we will accept.
func (a *agent) compile() {
	pattern := regexp.QuoteMeta(a.name)
	if pattern == `\*` {
		pattern = strings.Replace(pattern, `\*`, `.*`, -1)
	}
	pattern = "^" + pattern
	pattern = strings.ToLower(pattern)
	// FIXME: What do I do in case of error?
	r, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("oh no! %v\n", err)
	}
	a.pattern = r
}

// Robots represents the result of parsing a robots.txt file. To test
// whether an agent may crawl a path, use a Test* method. If any
// sitemaps were discovered while parsing, the Sitemaps field will be
// a slice containing their absolute URLs.
type Robots struct {
	// agents represents the groups of rules from a robots
	// file. The agents occur in descending order by length of
	// name. This ensures that if we check the agents
	// sequentially, the first matching agent will be the longest
	// match as well.
	agents   []*agent
	Sitemaps []string // Absolute URLs of sitemaps in robots.txt.
}

// Test takes an agent string and a rawurl string and checks whether the
// r allows name to access the path component of rawurl.
//
// Only the path of rawurl is used. For details, see method Tester.
func (r *Robots) Test(name, rawurl string) bool {
	return r.Tester(name)(rawurl)
}

// Tester takes string naming a user agent. It returns a predicate
// with a single string parameter representing a URL. This predicate
// can be used to check whether r allows name to crawl the path
// component of rawurl.
//
// Only the path part of rawurl is considered. Therefore, rawurl can
// be absolute or relative. It is the caller's responsibility to
// ensure that the Robots object is applicable to rawurl: no error can
// be provided if this is not the case. To ensure the Robots object is
// applicable to rawurl, use the Locate function.
func (r *Robots) Tester(name string) func(rawurl string) bool {
	agent, ok := r.bestAgent(name)
	if !ok {
		// An agent that isn't matched crawls everything.
		return func(_ string) bool {
			return true
		}
	}
	return func(rawurl string) bool {
		parsed, err := url.Parse(rawurl)
		if err != nil {
			return true
		}
		path := parsed.Path
		for _, member := range agent.group.members {
			if member.match(path) {
				return member.allow
			}
		}
		return true
	}
}

// addAgents adds a slice of agents to that maintained by r.
// This function accepts a slice because that is the common case:
// the parser may generate multiple agent objects from a single
// group of rules.
func (r *Robots) addAgents(agents []*agent) {
	for _, agent := range agents {
		// Maintain type invariant: all contained agents
		// must have patterns compiled before use.
		agent.compile()
	}
	r.agents = append(r.agents, agents...)
	// Maintain type invariant: r.agents must always be sorted
	// by length of agent name, descending.
	sort.Slice(r.agents, func(i, j int) bool {
		return len(r.agents[i].name) > len(r.agents[j].name)
	})
}

// bestAgent matches an agent string against all of the agents in
// r. It returns a pointer to the best matching agent, and a boolen
// indicating whether a match was found.
func (r *Robots) bestAgent(name string) (*agent, bool) {
	for _, agent := range r.agents {
		if agent.match(name) {
			return agent, true
		}
	}
	return nil, false
}
