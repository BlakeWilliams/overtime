package generator

import (
	goparser "go/parser"
	"go/token"
	"io"
	"testing"

	"github.com/blakewilliams/overtime/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestGoTypes(t *testing.T) {
	graph, err := parser.Parse(`
		Type Comment {
			id: int64
			body: string
		}`)

	require.NoError(t, err)

	gen := NewGo(graph)
	gen.PackageName = "mytypes"

	writer := gen.Types()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "ID int64 `json:\"id\"`")
	require.Contains(t, string(out), "Body string `json:\"body\"`")

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}

func TestGoResolvers(t *testing.T) {
	graph, err := parser.Parse(`
		Type Comment {
			id: int64
			body: string
		}
		Type Post {
			id: int64
			comments: []Comment
		}`)

	require.NoError(t, err)

	gen := NewGo(graph)
	gen.PackageName = "mytypes"

	writer := gen.Resolvers()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "type Resolver interface")
	require.Contains(t, string(out), "ResolvePostComments(postIDs []int64) map[int64]Comment")

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}
