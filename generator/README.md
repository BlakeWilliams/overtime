# Overtime

A simple DSL for defining federated (or non-federated) API's that auto-generate
type-safe, efficient APIs.

**Goals:**:

- Avoid writing boilerplate code for API's
- Make it easy to do the right thing, like auto-generating resolvers to avoid N+1 queries
- Generate type-safe clients for your API's
- Work with tools in the ecosystem, like openAPI.

## Usage

WIP, but here's an example of what the DSL looks like:

```
Type Comment {
  id: int64
  body: string
}

Type Post {
  fields {
    id: int64
    title: string
    body: string
    comments: []Comment
  }
}

Endpoint GET "/api/posts" {
  args {
    page?: int
  }

  fields {
    posts: []Post
  }
}
```

## TODO

- [ ] Finish Go auto-generation for resolvers and endpoints.
- [ ] Work on federation + enhancement capabilities by adding a gateway.
- [ ] Document the DSL and how to use it.
