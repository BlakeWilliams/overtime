package resolver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Post struct {
	ID       int64
	Comments []Comment `resolver:"ResolvePostComments"`
}

type Comment struct {
	ID   int64
	Body string
}

type TestResolver struct {
	// maps post id to comments for mock values
	commentMap map[int64][]Comment
}

func (r *TestResolver) ResolvePostComments(postIDs []int64) (map[int64][]Comment, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	res := make(map[int64][]Comment, len(postIDs))

	for _, id := range postIDs {
		res[id] = r.commentMap[id]
	}

	return res, nil
}

func TestResolve(t *testing.T) {
	resolver := &TestResolver{
		commentMap: map[int64][]Comment{
			1: {{ID: 1, Body: "This is a comment"}},
		},
	}

	post := &Post{
		ID: 1,
	}

	err := Resolve(post, resolver)
	require.NoError(t, err)

	require.Len(t, post.Comments, 1)
	require.Equal(t, int64(1), post.Comments[0].ID)
	require.Equal(t, "This is a comment", post.Comments[0].Body)
}

func TestResolve_Slice(t *testing.T) {
	resolver := &TestResolver{
		commentMap: map[int64][]Comment{
			1: {{ID: 1, Body: "This is a comment"}},
			2: {{ID: 2, Body: "This is a comment"}, {ID: 3, Body: "This is another comment"}},
		},
	}

	posts := []*Post{
		{ID: 1},
		{ID: 2},
	}

	err := Resolve(posts, resolver)
	require.NoError(t, err)

	firstPost := posts[0]
	secondPost := posts[1]

	require.Len(t, firstPost.Comments, 1)
	require.Len(t, secondPost.Comments, 2)
	require.Equal(t, int64(1), firstPost.Comments[0].ID)
	require.Equal(t, int64(2), secondPost.Comments[0].ID)
	require.Equal(t, int64(3), secondPost.Comments[1].ID)
}
