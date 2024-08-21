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
type Comment {
  id: int64
  body: string
}

type Post {
  fields {
    id: int64
    title: string
    body: string
    comments: []Comment
  }
}

GET "/api/posts" {
  input {
    page?: int
  }

  returns []Post
}
```

Endpoints can control what data type they return by specifying a `returns` block that is either an object, or an array.

**Array**:

```
  returns []Post
}
```

**Type**:

```
  fields Post
}
```

## TODO

- [ ] Finish Go auto-generation for resolvers and endpoints.
- [ ] Work on federation + enhancement capabilities by adding a gateway.
- [ ] Document the DSL and how to use it.
