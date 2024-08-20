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
		Partials  map[string]*Partial
	}

	// Endpoint represents a single endpoint in the schema. It is composed of
	// a path, partials, and fields.
	Endpoint struct {
		Path     string
		Method   string
		Partials []Partial
		Args     map[string]Field
		Fields   map[string]Field
	}

	// Partial represents a single partial in the schema. It is composed of a
	// name and a list of fields. It is the primary tool to keep consistency
	// within the schema.
	Partial struct {
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

func Parse(input string) (*Graph, error) {
	l := lexer.NewLexer(input)
	p := &parser{
		graph: &Graph{
			Endpoints: make(map[string]*Endpoint),
			Partials:  make(map[string]*Partial),
		},
	}

	err := p.parse(l)

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
			if t.Value == "Endpoint" {
				err := p.parseEndpoint(l)
				if err != nil {
					return err
				}
			} else if t.Value == "Partial" {
				err := p.parsePartial(l)
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

/*
	Endpoint /api/v1/comments {
	  args {
	    page?: int64
	  }

	  fields {
		comments []Comment
	  }
	}
*/
func (p *parser) parseEndpoint(l *lexer.Lexer) error {
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
				p.parseArgs(l, endpoint, true)
			case "fields":
				p.parseArgs(l, endpoint, false)
			default:
				l.PanicForToken(next, lexer.LexIdentifier)
			}
		}
	}

	p.graph.Endpoints[path] = endpoint

	return nil
}

func (p *parser) parseArgs(l *lexer.Lexer, endpoint *Endpoint, supportsOptional bool) {
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

		target := endpoint.Args
		if !supportsOptional {
			target = endpoint.Fields
		}
		target[name] = Field{
			Name:       name,
			Type:       identifier,
			IsOptional: optional,
		}

		expect(l, lexer.LexWhitespace)
	}
}

func (p *parser) parsePartial(l *lexer.Lexer) error {
	expect(l, lexer.LexWhitespace)
	name := expect(l, lexer.LexIdentifier).Value

	expect(l, lexer.LexWhitespace)
	expect(l, lexer.LexOpenCurly)

	partial := &Partial{
		Name:   name,
		Fields: make(map[string]Field),
	}

	for {
		expect(l, lexer.LexWhitespace)

		t := l.Next()
		if t.Kind == lexer.LexCloseCurly {
			break
		}

		if t.Kind != lexer.LexIdentifier {
			panic("unexpected token, expected identifier")
		}

		name := t.Value
		expect(l, lexer.LexWhitespace)
		fieldType := expect(l, lexer.LexIdentifier).Value

		partial.Fields[name] = Field{
			Name: name,
			Type: fieldType,
		}
	}

	p.graph.Partials[name] = partial

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
