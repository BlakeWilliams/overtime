package parser

import (
	"regexp"
	"strings"
	"sync"

	"github.com/blakewilliams/overtime/internal/lexer"
)

type (
	parser struct {
		graph       *Graph
		lastComment string
	}
)

func Parse(input string) (g *Graph, err error) {
	// The parser and lexer use panic to handle errors since maintaining errors
	// up the stack is extremely verbose and error prone. This ensures
	// that we handle errors gracefully upstream
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	l := lexer.NewLexer(input)
	p := &parser{
		graph: &Graph{
			Endpoints: make(map[string]*Endpoint),
			Types:     make(map[string]*Type),
		},
	}

	err = p.parse(l)

	for _, e := range p.graph.Endpoints {
		if err := e.Validate(); err != nil {
			return p.graph, err
		}
	}

	return p.graph, err
}

func (p *parser) parse(l *lexer.Lexer) error {
outer:
	for {
		t := l.Next()

		switch t.Kind {
		case lexer.LexEOF:
			break outer
		case lexer.LexWhitespace:
			continue
		case lexer.LexComment:
			p.lastComment = t.Value
			continue
		case lexer.LexIdentifier:
			switch t.Value {
			case "GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT":
				err := p.parseEndpoint(l, t.Value)
				if err != nil {
					return err
				}
			case "type":
				err := p.parseType(l)
				if err != nil {
					return err
				}
			}
		default:
			l.PanicForToken(t, lexer.LexIdentifier)
		}
	}

	p.lastComment = ""

	return nil
}

func (p *parser) consumeComment() string {
	c := sanitizeComment(p.lastComment)
	p.lastComment = ""
	return c
}

func (p *parser) parseEndpoint(l *lexer.Lexer, method string) (err error) {
	expect(l, lexer.LexWhitespace)
	path := expect(l, lexer.LexString).Value
	path = path[1 : len(path)-1]

	endpoint := &Endpoint{
		Path:       path,
		Method:     method,
		Args:       make(map[string]Field),
		DocComment: p.consumeComment(),
	}

	_, ok := p.graph.Endpoints[endpoint.Path]
	if ok {
		panic("Endpoint already exists")
	}

	expect(l, lexer.LexWhitespace)
	expect(l, lexer.LexOpenCurly)

outer:
	for {
		expectAtLeastOneWhitespace(l)
		next := l.Next()
		switch next.Kind {
		case lexer.LexCloseCurly:
			break outer
		case lexer.LexIdentifier:
			switch next.Value {
			case "input":
				expect(l, lexer.LexColon)
				p.parseFields(l, endpoint.Args, true)
			case "returns":
				p.parseEndpointReturn(l, endpoint)
			case "name":
				expect(l, lexer.LexColon)
				skipWhitespace(l)
				name := expect(l, lexer.LexIdentifier).Value
				endpoint.Name = name
			default:
				l.PanicForToken(next, lexer.LexIdentifier)
			}
		}
	}

	p.graph.Endpoints[path] = endpoint

	return nil
}

func (p *parser) parseEndpointReturn(l *lexer.Lexer, e *Endpoint) {
	expect(l, lexer.LexColon)
	skipWhitespace(l)

	next := l.Next()
	switch next.Kind {
	case lexer.LexOpenBracket:
		expect(l, lexer.LexCloseBracket)
		rawType := expect(l, lexer.LexIdentifier)

		e.Returns = "[]" + rawType.Value
	case lexer.LexIdentifier:
		e.Returns = next.Value
	default:
		l.PanicForToken(next, lexer.LexIdentifier)
	}
}

func (p *parser) parseFields(l *lexer.Lexer, target map[string]Field, supportsOptional bool) {
	expect(l, lexer.LexWhitespace)
	expect(l, lexer.LexOpenCurly)
	expect(l, lexer.LexWhitespace)

	for {
		t := l.Next()
		if t.Kind == lexer.LexCloseCurly {
			break
		}

		comment := ""
		if t.Kind == lexer.LexComment {
			comment = t.Value
			skipWhitespace(l)
			t = l.Next()
		}

		if t.Kind != lexer.LexIdentifier {
			l.PanicForToken(t, lexer.LexIdentifier)
		}

		name := t.Value

		next := l.Next()
		optional := false
		switch next.Kind {
		case lexer.LexQuestion:
			if !supportsOptional {
				l.PanicForToken(next, lexer.LexIdentifier)
			}
			optional = true
			expect(l, lexer.LexColon)
		case lexer.LexColon:
		default:
			l.PanicForToken(next, lexer.LexColon)
		}

		expectAtLeastOneWhitespace(l)

		identifier := ""
		switch l.Peek().Kind {
		case lexer.LexOpenBracket:
			identifier += expect(l, lexer.LexOpenBracket).Value
			identifier += expect(l, lexer.LexCloseBracket).Value
			identifier += expect(l, lexer.LexIdentifier).Value
		case lexer.LexIdentifier:
			identifier += expect(l, lexer.LexIdentifier).Value
		default:
			l.PanicForToken(l.Peek(), lexer.LexIdentifier)
		}

		target[name] = Field{
			Name:       name,
			Type:       identifier,
			IsOptional: optional,
			DocComment: sanitizeComment(comment),
		}

		expect(l, lexer.LexWhitespace)
	}
}

func (p *parser) parseType(l *lexer.Lexer) error {
	expect(l, lexer.LexWhitespace)
	name := expect(l, lexer.LexIdentifier).Value

	graphType := &Type{
		Name:       name,
		Fields:     make(map[string]Field),
		DocComment: p.consumeComment(),
	}

	p.parseFields(l, graphType.Fields, false)

	p.graph.Types[name] = graphType

	return nil
}

func expectAtLeastOneWhitespace(l *lexer.Lexer) {
	expect(l, lexer.LexWhitespace)
	skipWhitespace(l)
}

func skipWhitespace(l *lexer.Lexer) {
	for {
		switch l.Peek().Kind {
		case lexer.LexWhitespace:
			l.Next()
		case lexer.LexEOF:
			return
		default:
			return
		}
	}
}

func expect(l *lexer.Lexer, kind lexer.LexKind) lexer.Token {
	t := l.Next()
	if t.Kind == lexer.LexComment {
		t = l.Next()
	}
	if t.Kind != kind {
		l.PanicForToken(t, kind)
	}

	return t
}

var commentContentRegex = regexp.MustCompile(`^\s*#(.*?)(\r\n|\r|\n)?$`)

func sanitizeComment(comment string) string {
	if comment == "" {
		return ""
	}

	parts := strings.Split(comment, "\n")
	formatted := strings.Builder{}

	countWhitespace := sync.Once{}
	leadingWhitespace := 0

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			formatted.WriteString("\n\n")
			continue
		}

		if matches := commentContentRegex.FindStringSubmatch(part); len(matches) > 1 {
			content := matches[1]
			countWhitespace.Do(func() {
				leadingWhitespace = countLeadingWhitespace(content)
			})

			formatted.WriteString(content[leadingWhitespace:])
		}

	}

	return formatted.String()
}

func countLeadingWhitespace(s string) int {
	for i, r := range []rune(s) {
		if r != ' ' {
			return i
		}
	}

	return 0
}
