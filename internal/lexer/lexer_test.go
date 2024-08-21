package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLex(t *testing.T) {
	lexer := NewLexer(`Type Comment { id: int64 }`)

	requireToken(t, lexer.Next(), LexIdentifier, "Type")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "Comment")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexOpenCurly, "{")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "id")
	requireToken(t, lexer.Next(), LexColon, ":")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "int64")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexCloseCurly, "}")
}

func TestLexEndpoint(t *testing.T) {
	lexer := NewLexer(`Endpoint GET "/api/v1/comments" {
			args { page?: int64 }
			fields {
				comments: []Comment
			}
		}`,
	)

	requireToken(t, lexer.Next(), LexIdentifier, "Endpoint")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "GET")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexString, `"/api/v1/comments"`)
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexOpenCurly, "{")
	requireToken(t, lexer.Next(), LexWhitespace, "\n\t\t\t")
	requireToken(t, lexer.Next(), LexIdentifier, "args")

	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexOpenCurly, "{")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "page")
	requireToken(t, lexer.Next(), LexQuestion, "?")
	requireToken(t, lexer.Next(), LexColon, ":")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "int64")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexCloseCurly, "}")

	requireToken(t, lexer.Next(), LexWhitespace, "\n\t\t\t")
	requireToken(t, lexer.Next(), LexIdentifier, "fields")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexOpenCurly, "{")
	requireToken(t, lexer.Next(), LexWhitespace, "\n\t\t\t\t")
	requireToken(t, lexer.Next(), LexIdentifier, "comments")
	requireToken(t, lexer.Next(), LexColon, ":")

	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexOpenBracket, "[")
	requireToken(t, lexer.Next(), LexCloseBracket, "]")
	requireToken(t, lexer.Next(), LexIdentifier, "Comment")
	requireToken(t, lexer.Next(), LexWhitespace, "\n\t\t\t")
	requireToken(t, lexer.Next(), LexCloseCurly, "}")
	requireToken(t, lexer.Next(), LexWhitespace, "\n\t\t")
	requireToken(t, lexer.Next(), LexCloseCurly, "}")
	requireToken(t, lexer.Next(), LexEOF, "")
}

func TestComments(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		lexer := NewLexer(`# Hello world`)
		requireToken(t, lexer.Next(), LexComment, "# Hello world")
	})

	t.Run("leading space", func(t *testing.T) {
		lexer := NewLexer(` # Hello world`)
		requireToken(t, lexer.Next(), LexWhitespace, " ")
		requireToken(t, lexer.Next(), LexComment, "# Hello world")
	})

	t.Run("multi-line", func(t *testing.T) {
		lexer := NewLexer(" # Hello world\n# Whatsup?")
		requireToken(t, lexer.Next(), LexWhitespace, " ")
		requireToken(t, lexer.Next(), LexComment, "# Hello world\n# Whatsup?")
	})

	t.Run("multi-line with content", func(t *testing.T) {
		lexer := NewLexer(" # Hello world\n# Whatsup?\n type Foo {}")
		requireToken(t, lexer.Next(), LexWhitespace, " ")
		requireToken(t, lexer.Next(), LexComment, "# Hello world\n# Whatsup?")
		requireToken(t, lexer.Next(), LexWhitespace, "\n ")
		requireToken(t, lexer.Next(), LexIdentifier, "type")
	})
}

func requireToken(t testing.TB, input Token, kind LexKind, value string) {
	require.Equal(t, kind, input.Kind)
	require.Equal(t, value, input.Value)
}
