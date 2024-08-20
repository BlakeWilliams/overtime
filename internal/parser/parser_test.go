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
			input: `Partial Comment { id int64 }`,
			expectedGraph: &Graph{
				Endpoints: map[string]*Endpoint{},
				Partials: map[string]*Partial{
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
				Endpoint GET "/api/v1/comments" {
					args { page?: int64 }
					fields {
						comments: []Comment
					}
				}`,

			expectedGraph: &Graph{
				Partials: map[string]*Partial{},
				Endpoints: map[string]*Endpoint{
					"/api/v1/comments": {
						Method: "GET",
						Path:   "/api/v1/comments",
						Args: map[string]Field{
							"page": {
								Name:       "page",
								Type:       "int64",
								IsOptional: true,
							},
						},
						Fields: map[string]Field{
							"comments": {
								Name: "comments",
								Type: "[]Comment",
							},
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			graph, err := Parse(tC.input)
			require.NoError(t, err)

			for name, partial := range graph.Partials {
				require.Equal(t, tC.expectedGraph.Partials[name], partial)
			}

			for path, endpoint := range graph.Endpoints {
				require.Equal(t, tC.expectedGraph.Endpoints[path], endpoint)
			}
		})
	}
}
