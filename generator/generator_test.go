package generator

import (
	goparser "go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/blakewilliams/overtime/internal/parser"
	"github.com/stretchr/testify/require"
)

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
	require.Contains(t, string(out), "ResolvePostComments(postIDs []int64) (map[int64][]*Comment, error)")

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

func TestCoordinator(t *testing.T) {
	graph, err := parser.Parse(`
		type Comment {
			id: int64
			body: string
		}

		type Post {
			id: int64
			body: string
			comments: []Comment
		}

		GET "/api/v1/comments/:commentID" {
			name: GetCommentByID
			returns: Comment
		}

		GET "/api/v1/posts/:postID" {
			name: GetPostByID
			returns: Post
		}`)

	require.NoError(t, err)

	gen := NewGo(graph)
	gen.PackageName = "mytypes"

	writer := gen.Coordinator()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "type Coordinator struct")
	require.Contains(t, string(out), "GET /api/v1/comments/{commentID}")
	require.Contains(t, string(out), "result, err := c.controller.GetCommentByID(w, r)")
	require.Contains(t, string(out), "ResolveForPost([]*Post{result}, c.resolver)")

	// controller tests
	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "GetCommentByID(w http.ResponseWriter, r *http.Request)")

	// type tests
	require.Regexp(t, regexp.MustCompile("ID\\s+int64\\s+`json:\"id\"`"), string(out))
	require.Regexp(t, regexp.MustCompile("Comments\\s+\\[\\]\\*Comment\\s+`json:\"comments\" resolver:\"ResolvePostComments\""), string(out))

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}

func Test_EndToEnd(t *testing.T) {
	cmd := exec.Command("go", "run", "../main.go", "generate", "./e2e.ovt", "-d", "generator/test")
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	require.NoError(t, err)
}
