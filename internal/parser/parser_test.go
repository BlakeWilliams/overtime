package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		desc          string
		input         string
		expectedGraph *Graph
	}{
		{
			desc:  "basic partial",
			input: `type Comment { id: int64 }`,
			expectedGraph: &Graph{
				Endpoints: map[string]*Endpoint{},
				Types: map[string]*Type{
					"Comment": {
						Name: "Comment",
						Fields: map[string]Field{
							"id": {
								Name: "id",
								Type: "int64",
							},
						},
					},
				},
			},
		},
		{
			desc: "basic endpoint",
			input: `
				GET "/api/v1/comments" {
					name: GetComments
					input: { page?: int64 }
					returns: []Comment
				}`,

			expectedGraph: &Graph{
				Types: map[string]*Type{},
				Endpoints: map[string]*Endpoint{
					"/api/v1/comments": {
						Method: "GET",
						Name:   "GetComments",
						Path:   "/api/v1/comments",
						Args: map[string]Field{
							"page": {
								Name:       "page",
								Type:       "int64",
								IsOptional: true,
							},
						},
						Returns: "[]Comment",
					},
				},
			},
		},
		{
			desc: "basic endpoint with comments",
			input: `
				# GetComments returns a list of comments
				GET "/api/v1/comments" { # Shouldn't break
					name: GetComments # shouldn't break either
					input: {
						# page is the page number
						page?: int64
					}
					returns: []Comment
				}`,

			expectedGraph: &Graph{
				Types: map[string]*Type{},
				Endpoints: map[string]*Endpoint{
					"/api/v1/comments": {
						DocComment: "GetComments returns a list of comments",
						Method:     "GET",
						Name:       "GetComments",
						Path:       "/api/v1/comments",
						Args: map[string]Field{
							"page": {
								Name:       "page",
								Type:       "int64",
								IsOptional: true,
								DocComment: "page is the page number",
							},
						},
						Returns: "[]Comment",
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			graph, err := Parse(tC.input)
			require.NoError(t, err)

			for name, partial := range graph.Types {
				require.Equal(t, tC.expectedGraph.Types[name], partial)
			}

			for path, endpoint := range graph.Endpoints {
				require.Equal(t, tC.expectedGraph.Endpoints[path], endpoint)
			}
		})
	}
}
