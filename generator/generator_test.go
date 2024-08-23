package generator

import (
	goparser "go/parser"
	"go/token"
	"io"
	"regexp"
	"testing"

	"github.com/blakewilliams/overtime/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestGoTypes(t *testing.T) {
	graph, err := parser.Parse(`
		# A comment on a post
		type Comment {
			# The ID of the comment
			id: int64
			# The body of the comment
			body: string
		}
			
		type Post {
			id: int64
			comments: []Comment
		}`)

	require.NoError(t, err)

	gen := NewGo(graph)
	gen.PackageName = "mytypes"

	writer := gen.Types()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Regexp(t, regexp.MustCompile("ID\\s+int64\\s+`json:\"id\"`"), string(out))
	require.Regexp(t, regexp.MustCompile("Comments\\s+\\[\\]Comment\\s+`json:\"comments\" resolver:\"ResolvePostComments\""), string(out))
	// require.Regexp(t, regexp.MustCompile("Comments\\s+\\[\\]Comment\\s+`json:\"comments\" resolver:\"`"), string(out))

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}

func TestGoResolvers(t *testing.T) {
	graph, err := parser.Parse(`
		type Comment {
			id: int64
			body: string
		}
		type Post {
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
	require.Contains(t, string(out), "ResolvePostComments(postIDs []int64) (map[int64]Comment, error)")

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}

func TestGoEndpoints(t *testing.T) {
	graph, err := parser.Parse(`
		type Comment {
			id: int64
			body: string
		}

		GET "/api/v1/comments/:commentID" {
			name: GetCommentByID
			returns: Comment
		}`)

	require.NoError(t, err)

	gen := NewGo(graph)
	gen.PackageName = "mytypes"

	writer := gen.Endpoints()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "type Controller interface")
	require.Contains(t, string(out), "GetCommentByID(w http.ResponseWriter, r *http.Request)")

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}

func TestRoot(t *testing.T) {
	graph, err := parser.Parse(``)

	require.NoError(t, err)

	gen := NewGo(graph)
	gen.PackageName = "mytypes"

	writer := gen.Root()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "type RootResolver")
	require.Contains(t, string(out), "type RootController")

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}
