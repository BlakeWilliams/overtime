package parser

import (
	"github.com/blakewilliams/overtime/internal/lexer"
)

type (
	// Graph represents the entire schema for all federated services and is
	// responsible for combining the relevant schemas of each service into a
	// single schema that can be used to generate a gateway.
	Graph struct {
		Endpoints map[string]*Endpoint
		Types     map[string]*Type
	}

	// Endpoint represents a single endpoint in the schema. It is composed of
	// a path, types, and fields.
	Endpoint struct {
		Path   string
		Method string
		Args   map[string]Field
		Fields map[string]Field
	}

	// Type represents a single partial in the schema. It is composed of a
	// name and a list of fields. It is the primary tool to keep consistency
	// within the schema.
	Type struct {
		Name   string
		Fields map[string]Field
		// TODO fit in federation pieces here
	}

	// Field represents a single field in the schema. It is composed of a name
	// and a type. The type is a string that represents the type of the field.
	// This is a string because the type could be a scalar, an object, or a
	// list of objects.
	Field struct {
		Name       string
		Type       string
		IsOptional bool
		IsPartial  bool
	}

	parser struct {
		graph *Graph
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
			continue
		case lexer.LexIdentifier:
			switch t.Value {
			case "Endpoint":
				err := p.parseEndpoint(l)
				if err != nil {
					return err
				}
			case "Type":
				err := p.parseType(l)
				if err != nil {
					return err
				}
			}
		default:
			if t.Kind == "" {
				break
			}
			l.PanicForToken(t, lexer.LexIdentifier)
		}
	}

	return nil
}

func (p *parser) parseEndpoint(l *lexer.Lexer) (err error) {
	expect(l, lexer.LexWhitespace)
	method := expect(l, lexer.LexIdentifier).Value

	expect(l, lexer.LexWhitespace)
	path := expect(l, lexer.LexString).Value
	path = path[1 : len(path)-1]

	endpoint := &Endpoint{
		Path:   path,
		Method: method,
		Args:   make(map[string]Field),
		Fields: make(map[string]Field),
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
			case "args":
				p.parseFields(l, endpoint.Args, true)
			case "fields":
				p.parseFields(l, endpoint.Fields, false)
			default:
				l.PanicForToken(next, lexer.LexIdentifier)
			}
		}
	}

	p.graph.Endpoints[path] = endpoint

	return nil
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
		}

		expect(l, lexer.LexWhitespace)
	}
}

func (p *parser) parseType(l *lexer.Lexer) error {
	expect(l, lexer.LexWhitespace)
	name := expect(l, lexer.LexIdentifier).Value

	graphType := &Type{
		Name:   name,
		Fields: make(map[string]Field),
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
	if t.Kind != kind {
		l.PanicForToken(t, kind)
	}

	return t
}
