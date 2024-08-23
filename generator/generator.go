package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/blakewilliams/overtime/internal/parser"
)

var builtins = map[string]bool{
	"int":     true,
	"int64":   true,
	"string":  true,
	"bool":    true,
	"float":   true,
	"float64": true,
}

type Go struct {
	graph       *parser.Graph
	PackageName string
}

func (g *Go) Endpoints() []Endpoint {
	endpoints := make([]Endpoint, 0, len(g.graph.Endpoints))
	for _, e := range g.graph.Endpoints {
		endpoints = append(endpoints, Endpoint{endpoint: e, graph: g.graph})
	}

	return endpoints
}

func (g *Go) Types() []GoType {
	types := make([]GoType, 0, len(g.graph.Types))
	for _, t := range g.graph.Types {
		types = append(types, GoType{parserType: t})
	}

	return types
}

func NewGo(graph *parser.Graph) *Go {
	return &Go{graph: graph, PackageName: "types"}
}

func (g *Go) Root() io.Reader {
	buf := new(bytes.Buffer)

	buf.WriteString("// This file is generated only once to bootstrap the project\n")
	buf.WriteString("// Your implementation for resolvers and endpoints should go here\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	buf.WriteString("type RootResolver struct {}\n\n")
	buf.WriteString("var _ Resolver = (*RootResolver)(nil)\n\n")
	buf.WriteString("type RootController struct {}\n\n")
	buf.WriteString("var _ Controller = (*RootController)(nil)\n\n")

	return formatCode(buf)
}

func (g *Go) Coordinator() io.Reader {
	template, err := template.New("server").Parse(`
	package {{.PackageName}}

	import (
		"net/http"
		"encoding/json"
	)

	// Coordinator is the main entrypoint for the server and is responsible for
	// routing requests to the correct endpoint and invoking the correct method
	// on the controller. It also handles serializing the response and calling
	// resolver methods to efficiently fetch related data.
	type Coordinator struct {
		mux 		http.ServeMux
		resolver 	Resolver
		controller 	Controller
	}

	// NewCoordinator returns a new Coordinator that passes requests to the
	// provided resolver and controller.
	func NewCoordinator(resolver Resolver, controller Controller) *Coordinator {
		c := &Coordinator{
			mux: http.ServeMux{},
			resolver: resolver,
			controller: controller,
		}

		{{ range $key, $value := .Endpoints }}
		c.mux.HandleFunc("{{.Method }} {{.Path}}", func(w http.ResponseWriter, r *http.Request) {
			result, err := c.controller.{{ .MethodName }}(w, r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			{{ if .ResolverMethod }}
				{{ .ResolverMethod }}
			{{ end }}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(result)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
		{{ end }}

		return c
	}

	// ServeHTTP serves the provided request by routing it to the correct
	// endpoint and invoking the correct method on the controller.
	func (c *Coordinator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		c.mux.ServeHTTP(w, r)
	}
	`)

	if err != nil {
		panic(fmt.Errorf("failed to generate server template: %w", err))
	}

	buf := new(bytes.Buffer)

	buf.WriteString("// Code generated by github.com/blakewilliams/overtime DO NOT EDIT\n\n")

	err = template.Execute(buf, map[string]interface{}{
		"PackageName": g.PackageName,
		"Endpoints":   g.Endpoints(),
	})

	if err != nil {
		panic(fmt.Errorf("failed to execute server template: %w", err))
	}

	return formatCode(buf)
}

func (g *Go) GenerateControllers() io.Reader {
	buf := new(bytes.Buffer)

	buf.WriteString("// Code generated by github.com/blakewilliams/overtime DO NOT EDIT\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	buf.WriteString("import (\n\t\"net/http\"\n)\n\n")

	buf.WriteString("type Controller interface {\n")
	for _, e := range g.Endpoints() {
		buf.WriteString(
			fmt.Sprintf(
				"\t%s(w http.ResponseWriter, r *http.Request) (%s, error)\n",
				e.MethodName(),
				e.ReturnValue(),
			),
		)
	}

	buf.WriteString("}\n")

	return formatCode(buf)
}

func (g *Go) GenerateTypes() io.Reader {
	buf := new(bytes.Buffer)

	buf.WriteString("// Code generated by github.com/blakewilliams/overtime DO NOT EDIT\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	for _, t := range g.Types() {
		buf.WriteString(fmt.Sprintf("type %s struct {\n", t.Name()))

		for _, field := range t.Fields() {
			buf.WriteString(fmt.Sprintf("\t%s %s%s\n", field.Name(), field.Type(), field.Tags()))
		}

		buf.WriteString("}\n\n")

		if t.NeedsResolver() {
			// Now that the type is defined, we need to define the resolver method
			buf.WriteString(
				fmt.Sprintf(
					"func ResolveFor%s(records []*%s, resolver Resolver) (error) {\n",
					capitalize(t.Name()),
					capitalize(t.Name()),
				),
			)

			buf.WriteString(fmt.Sprintf(`ids := make([]int64, len(records))
			recordsMap := make(%s, len(records))
			for i, record := range records {
				ids[i] = record.ID
				recordsMap[record.ID] = record
			}
		`, t.MapType()))

			for _, field := range t.Fields() {
				if builtins[field.normalizedType()] {
					continue
				}

				buf.WriteString(fmt.Sprintf(`{
				res, err := resolver.%s(ids)
				if err != nil {
					return err
				}

				for id, record := range recordsMap {
					if val, ok := res[id]; ok {
						record.%s = val
					}
				}
			}`,
					field.ResolverMethodName(),
					field.Name(),
				))
				buf.WriteRune('\n')
			}

			buf.WriteString("\treturn nil\n}\n\n")
		}
	}

	return formatCode(buf)
}

func (g *Go) Resolvers() io.Reader {
	buf := new(bytes.Buffer)
	buf.WriteString("// Code generated by github.com/blakewilliams/overtime DO NOT EDIT\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	addedResolvers := map[string]bool{}

	buf.WriteString("type Resolver interface {\n")

	for _, t := range g.graph.Types {
		for _, f := range t.Fields {
			normalizedType := strings.TrimPrefix(f.Type, "[]")
			if addedResolvers[normalizedType] || builtins[normalizedType] {
				continue
			}

			customType := g.graph.Types[normalizedType]
			if customType == nil {
				continue
			}

			// TODO panic if id field is not present
			idType := t.Fields["id"].Type

			arguments := fmt.Sprintf(
				"%sIDs []%s",
				uncapitalize(t.Name),
				idType,
			)

			mapType := ""
			if strings.HasPrefix(f.Type, "[]") {
				mapType = "[]*" + capitalize(strings.TrimPrefix(f.Type, "[]"))
			} else {
				mapType = "*" + capitalize(f.Type)
			}
			buf.WriteString(
				fmt.Sprintf(
					"\tResolve%s%s(%s) (map[%s]%s, error)\n",
					capitalize(t.Name),
					capitalize(f.Name),
					arguments,
					idType,
					mapType,
				),
			)
		}
	}

	buf.WriteString("}\n\n")

	return formatCode(buf)
}

func uncapitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)

	return string(append([]rune{unicode.ToLower(r[0])}, r[1:]...))
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)

	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}

func formatCode(b *bytes.Buffer) io.Reader {
	formatted, err := format.Source(b.Bytes())
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(formatted)
}
