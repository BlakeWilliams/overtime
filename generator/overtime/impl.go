// This file is generated only once to bootstrap the project
// Your implementation for resolvers and endpoints should go here
package overtime

import "net/http"

type RootResolver struct{}

var _ Resolver = (*RootResolver)(nil)

func (r *RootResolver) ResolvePostComments(ids []int64) (map[int64][]*Comment, error) {
	return map[int64][]*Comment{
		1: {{ID: 1, Body: "comment 1"}, {ID: 2, Body: "comment 2"}},
	}, nil
}

type RootController struct{}

var _ Controller = (*RootController)(nil)

func (c *RootController) GetCommentByID(w http.ResponseWriter, r *http.Request) (*Comment, error) {
	return &Comment{
		ID:   1,
		Body: "comment 1",
	}, nil
}

func (c *RootController) GetPostByID(w http.ResponseWriter, r *http.Request) (*Post, error) {
	return &Post{
		ID:   1,
		Body: "post 1",
	}, nil
}
