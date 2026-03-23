package mysql

import (
	"fmt"

	"github.com/vjeantet/grok"
)

type SlowLogParser struct {
	pattern map[string]string
	g       *grok.Grok
}

func NewSlowLogParser(pattern map[string]string) (parser *SlowLogParser, err error) {
	var g *grok.Grok

	if pattern == nil {
		return nil, fmt.Errorf("pattern is nil")
	}

	if g, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true, RemoveEmptyValues: true}); err != nil {
		return nil, err
	}

	if err := g.AddPatternsFromMap(pattern); err != nil {
		return nil, err
	}

	return &SlowLogParser{
		pattern: pattern,
		g:       g,
	}, nil
}

func (p *SlowLogParser) ParserSlowLog(log string) (map[string]interface{}, error) {
	return p.g.ParseTyped("%{SLOW}", log)
}
