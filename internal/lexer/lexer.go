package lexer

import (
	"fmt"
	"unicode"
)

type LexKind string

const (
	LexEOF          LexKind = "eof"
	LexComment              = "comment"
	LexIdentifier           = "identifier"
	LexWhitespace           = "whitespace"
	LexDash                 = "dash"
	LexColon                = "colon"
	LexOpenCurly            = "open_curly"
	LexCloseCurly           = "close_curly"
	LexQuestion             = "question"
	LexOpenBracket          = "open_bracket"
	LexCloseBracket         = "close_bracket"
	LexSlash                = "slash"
	LexString               = "string"
)

type Token struct {
	Kind     LexKind
	Value    string
	StartPos int
	EndPos   int
}

type Lexer struct {
	Input    []rune
	Pos      int
	StartPos int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		Input: []rune(input),
	}
}

func (l *Lexer) emit(kind LexKind) Token {
	start := l.StartPos
	l.StartPos = l.Pos

	return Token{Kind: kind, Value: string(l.Input[start:l.Pos]), StartPos: start, EndPos: l.Pos}
}

func (l *Lexer) nextChar() rune {
	r := l.Input[l.Pos]
	l.Pos++
	return r
}

func (l *Lexer) Peek() Token {
	t := l.Next()
	l.StartPos = t.StartPos
	l.Pos = t.StartPos
	return t
}

func (l *Lexer) Backup() {
	l.Pos--
}

func (l *Lexer) Next() Token {
	if l.Pos >= len(l.Input) {
		return Token{Kind: LexEOF, Value: "", StartPos: l.Pos, EndPos: l.Pos}
	}

	ch := l.nextChar()
	switch ch {
	case ' ', '\t', '\r', '\n':
		return l.emitWhitespace()
	case '-':
		return l.emit(LexDash)
	case ':':
		return l.emit(LexColon)
	case '?':
		return l.emit(LexQuestion)
	case '{':
		return l.emit(LexOpenCurly)
	case '}':
		return l.emit(LexCloseCurly)
	case '[':
		return l.emit(LexOpenBracket)
	case ']':
		return l.emit(LexCloseBracket)
	case '/':
		return l.emit(LexSlash)
	default:
		if unicode.IsLetter(rune(ch)) {
			return l.emitIdentifier()
		} else if ch == '"' {
			return l.emitString()
		} else {
			panic(fmt.Sprintf("Unexpected character: %s", string(ch)))
		}
	}
}

func (l *Lexer) PanicForToken(t Token, expected LexKind) {
	line := 1
	column := 1

	// this could be more efficient, but it's only in the panic case so it's fine for now
	for i := 0; i < t.StartPos; i++ {
		if l.Input[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	panic(
		fmt.Sprintf(
			"Unexpected token type %s with value `%s` on line %d, columns %d, expected %s",
			t.Kind,
			t.Value,
			line,
			column,
			expected,
		),
	)
}

func (l *Lexer) emitWhitespace() Token {
	for {
		if l.Pos >= len(l.Input) {
			break
		}

		if l.Input[l.Pos] != ' ' && l.Input[l.Pos] != '\t' && l.Input[l.Pos] != '\r' && l.Input[l.Pos] != '\n' {
			break
		}

		l.Pos++
	}

	return l.emit(LexWhitespace)
}

func (l *Lexer) emitIdentifier() Token {
	for l.Pos < len(l.Input) && (unicode.IsLetter(rune(l.Input[l.Pos])) || unicode.IsDigit(rune(l.Input[l.Pos]))) {
		l.Pos++
	}

	return l.emit(LexIdentifier)
}

func (l *Lexer) emitString() Token {
	for {
		if l.Pos >= len(l.Input) {
			panic(fmt.Sprintf("Unterminated string starting at %d", l.StartPos))
		}

		next := l.nextChar()
		if next == '"' {
			break
		}
		if next == '\\' {
			l.nextChar()
		}
	}

	return l.emit(LexString)
}
