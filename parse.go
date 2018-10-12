package robots

import "sort"

type parser struct {
	agents      []*agent
	withinGroup bool
	items       []*item
	robots      robots
}

type parsefn func(p *parser) parsefn

func parse(s string) robots {
	items := lex(s)
	p := &parser{
		items: items,
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
	default:
		return parseNext
	}
}

func parseUserAgent(p *parser) parsefn {
	if p.withinGroup {
		p.appendAgents()
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

func (p *parser) appendAgents() {
	p.robots = append(p.robots, p.agents...)
	sort.Slice(p.robots, func(i, j int) bool {
		return len(p.robots[i].name) > len(p.robots[j].name)
	})
}

func parseDisallow(p *parser) parsefn {
	p.withinGroup = true
	for _, agent := range p.agents {
		m := &member{
			allow: false,
			path:  p.items[0].val,
		}
		m.compile()
		agent.group = append(agent.group, m)
		sort.Slice(agent.group, func(i, j int) bool {
			return len(agent.group[i].path) > len(agent.group[j].path)
		})
	}
	return parseNext
}

func parseAllow(p *parser) parsefn {
	p.withinGroup = true
	for _, agent := range p.agents {
		m := &member{
			allow: true,
			path:  p.items[0].val,
		}
		m.compile()
		agent.group = append(agent.group, m)
		sort.Slice(agent.group, func(i, j int) bool {
			return len(agent.group[i].path) > len(agent.group[j].path)
		})
	}
	return parseNext
}

func parseNext(p *parser) parsefn {
	p.items = p.items[1:]
	if len(p.items) == 0 {
		return parseEnd
	}
	return parseStart
}

func parseEnd(p *parser) parsefn {
	p.appendAgents()
	sort.Slice(p.robots, func(i, j int) bool {
		return len(p.robots[i].name) > len(p.robots[j].name)
	})
	for _, agent := range p.robots {
		agent.compile()
	}
	return nil
}
