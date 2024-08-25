package generator

import (
	goparser "go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/blakewilliams/overtime/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestCodeGen(t *testing.T) {
	schema, err := parser.Parse(strings.NewReader(`
types:
    Comment:
        fields:
            id:	int64
            body: string
    Post:
        fields:
            id:	int64
            body: string
            comments: "[]Comment"
endpoints:
    "GET /api/v1/comments/:commentID":
        name: GetCommentByID
        response:
            status: 200
            body: Comment
    "GET /api/v1/posts/:postID":
        name: GetPostByID
        response:
            status: 200
            body: Post`))

	require.NoError(t, err)

	gen := NewGo(schema)
	gen.PackageName = "mytypes"

	writer := gen.Coordinator()
	out, err := io.ReadAll(writer)
	require.NoError(t, err)

	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "type Coordinator struct")
	require.Contains(t, string(out), "GET /api/v1/comments/{commentID}")
	require.Contains(t, string(out), "result, err := c.controller.GetCommentByID(w, r)")
	require.Contains(t, string(out), "ResolveForPost([]*Post{result}, c.resolver)")

	// HandleFunc
	require.Contains(t, string(out), `HandleFunc("GET /api/v1/comments/{commentID}", func(w http.ResponseWriter, r *http.Request)`)
	require.Contains(t, string(out), "GetCommentByID(w http.ResponseWriter, r *http.Request)")

	// controller tests
	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "GetCommentByID(w http.ResponseWriter, r *http.Request)")

	// type tests
	require.Regexp(t, regexp.MustCompile("ID\\s+int64\\s+`json:\"id\"`"), string(out))
	require.Regexp(t, regexp.MustCompile("Comments\\s+\\[\\]\\*Comment\\s+`json:\"comments\" resolver:\"ResolvePostComments\""), string(out))

	// resolver tests
	require.Contains(t, string(out), `package mytypes`)
	require.Contains(t, string(out), "type Resolver interface")
	require.Contains(t, string(out), "ResolvePostComments(postIDs []int64) (map[int64][]*Comment, error)")

	fset := token.NewFileSet()
	_, err = goparser.ParseFile(fset, "", out, goparser.AllErrors)
	require.NoError(t, err, "Generated code should parse without errors")
}

func Test_EndToEnd(t *testing.T) {
	cmd := exec.Command("go", "run", "../main.go", "generate", "./e2e.yaml", "-d", "generator/test")
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	require.NoError(t, err)
}
