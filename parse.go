package robots

type parser struct {
	agents      []*agent
	withinGroup bool
	items       []*item
	robots      *Robots
}

type parsefn func(p *parser) parsefn

func parse(s string) *Robots {
	items := lex(s)
	p := &parser{
		items:  items,
		robots: &Robots{},
	}
	for fn := parseStart; fn != nil; fn = fn(p) {
	}
	return p.robots
}

func parseStart(p *parser) parsefn {
	switch p.items[0].typ {
	case itemUserAgent:
		return parseUserAgent
	case itemDisallow:
		return parseDisallow
	case itemAllow:
		return parseAllow
	case itemSitemap:
		return parseSitemap
	default:
		return parseNext
	}
}

func parseUserAgent(p *parser) parsefn {
	if p.withinGroup {
		p.robots.addAgents(p.agents)
		p.agents = nil
		p.agents = append(p.agents, &agent{
			name: p.items[0].val,
		})
		p.withinGroup = false
		return parseNext
	}
	p.agents = append(p.agents, &agent{
		name: p.items[0].val,
	})
	return parseNext
}

func makeParseMember(allow bool) func(*parser) parsefn {
	return func(p *parser) parsefn {
		p.withinGroup = true
		for _, agent := range p.agents {
			m := &member{
				allow: allow,
				path:  p.items[0].val,
			}
			agent.group.addMember(m)
		}
		return parseNext
	}
}

var parseDisallow parsefn

var parseAllow parsefn

func init() {
	parseDisallow = makeParseMember(false)
	parseAllow = makeParseMember(true)
}

func parseNext(p *parser) parsefn {
	p.items = p.items[1:]
	if len(p.items) == 0 {
		return parseEnd
	}
	return parseStart
}

func parseEnd(p *parser) parsefn {
	p.robots.addAgents(p.agents)
	return nil
}

func parseSitemap(p *parser) parsefn {
	p.robots.Sitemaps = append(p.robots.Sitemaps, p.items[0].val)
	return parseNext
}
