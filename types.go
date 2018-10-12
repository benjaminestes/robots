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
	// Following temoto's example.
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
// whether, under the rules of that file, an agent should crawl a path,
// use a Test* method. If an sitemaps were discovered in the robots.txt
// file, their absolute URLs are in the Robots.Sitemaps field.
type Robots struct {
	// agents represents the groups of rules from a robots
	// file. The agents occur in descending order by length of
	// name. This ensures that if we check the agents
	// sequentially, the first matching agent will be the longest
	// match as well.
	agents   []*agent
	Sitemaps []string // Absolute URLs of sitemaps in robots.txt.
}

// Test takes an agent string and a path string and checks whether the
// robots.txt file represented by r allows the named agent to crawl
// the named path.
//
// For details, see method Tester.
func (r *Robots) Test(a, p string) bool {
	return r.Tester(a)(p)
}

// Tester takes an agent string. It returns a predicate with a single
// string argument representing a path. This predicate can be used to
// check whether, under the robots.txt file represented by r, the
// agent a can crawl the path p.
//
// Only the path component of the provided URL is used. Therefore,
// input can be either an absolute or relative URL. It is the caller's
// responsibility to ensure that the Robots object is applicable
// to the URL in question: the scheme and domain of the raw URL
// will be discarded without warning. To ensure the Robots object
// is applicable to the raw URL, use the Locate function.
func (r *Robots) Tester(a string) func(p string) bool {
	agent, ok := r.bestAgent(a)
	if !ok {
		// An agent that isn't matched crawls everything.
		return func(_ string) bool {
			return true
		}
	}
	return func(path string) bool {
		parsed, err := url.Parse(path)
		if err != nil {
			return true
		}
		path = parsed.Path
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
func (r *Robots) bestAgent(a string) (*agent, bool) {
	for _, agent := range r.agents {
		if agent.match(a) {
			return agent, true
		}
	}
	return nil, false
}
