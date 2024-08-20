package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLex(t *testing.T) {
	lexer := NewLexer(`Partial Comment { id int64 }`)

	requireToken(t, lexer.Next(), LexIdentifier, "Partial")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "Comment")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexOpenCurly, "{")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "id")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexIdentifier, "int64")
	requireToken(t, lexer.Next(), LexWhitespace, " ")
	requireToken(t, lexer.Next(), LexCloseCurly, "}")
}

func TestLexEndpoint(t *testing.T) {
	lexer := NewLexer(`Endpoint GET "/api/v1/comments" {
			args { page?: int64 }
			fields {
				comments []Comment
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

func requireToken(t testing.TB, input Token, kind LexKind, value string) {
	require.Equal(t, kind, input.Kind)
	require.Equal(t, value, input.Value)
}
