// Parser takes a YAML file and returns the relevant internal representation of
// the schema.
package graph

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	// Schema represents the entire schema for all federated services and is
	// responsible for combining the relevant schemas of each service into a
	// single schema that can be used to generate a gateway.
	Schema struct {
		Endpoints map[string]*Endpoint
		Types     map[string]*Type
	}

	// Endpoint represents a single endpoint in the schema. It is composed of
	// a path, types, and fields.
	Endpoint struct {
		Name       string
		Path       string
		Method     string
		Args       map[string]Field
		Returns    string
		DocComment string
	}

	// Type represents a single partial in the schema. It is composed of a
	// name and a list of fields. It is the primary tool to keep consistency
	// within the schema.
	Type struct {
		Name       string
		Fields     map[string]Field
		DocComment string
		// TODO fit in federation pieces here
	}

	// Field represents a single field in the schema. It is composed of a name
	// and a type. The type is a string that represents the type of the field.
	// This is a string because the type could be a scalar, an object, or a
	// list of objects.
	Field struct {
		Name       string
		Type       string
		IsOptional bool
		IsPartial  bool
		DocComment string
	}

	// RawSchema is the representation of the raw schema file before converted
	// into an internal schema
	rawSchema struct {
		Types     map[string]rawType     `yaml:"types"`
		Endpoints map[string]rawEndpoint `yaml:"endpoints"`
	}

	rawEndpoint struct {
		Name     string      `yaml:"name"`
		Response rawResponse `yaml:"response"`
	}

	rawResponse struct {
		Status int    `yaml:"status"`
		Body   string `yaml:"body"`
	}

	rawType struct {
		Fields map[string]string `yaml:"fields"`
	}
)

func (e *Endpoint) Validate() error {
	if e.Path == "" {
		return fmt.Errorf("Path is required")
	}

	if e.Method == "" {
		return fmt.Errorf("Method is required")
	}

	if e.Name == "" {
		return fmt.Errorf("`name` is not defined for %s %s", e.Method, e.Path)
	}

	if e.Returns == "" {
		return fmt.Errorf("`returns` is not defined for %s %s", e.Method, e.Path)
	}

	return nil
}

func Parse(s io.Reader) (*Schema, error) {
	root := rawSchema{}
	err := yaml.NewDecoder(s).Decode(&root)
	if err != nil {
		return nil, err
	}

	schema := &Schema{
		Endpoints: make(map[string]*Endpoint, len(root.Endpoints)),
		Types:     make(map[string]*Type, len(root.Types)),
	}

	for name, rawType := range root.Types {
		t := &Type{
			Name:   name,
			Fields: make(map[string]Field, len(rawType.Fields)),
		}

		for fieldName, fieldType := range rawType.Fields {
			t.Fields[fieldName] = Field{
				Name: fieldName,
				Type: fieldType,
			}
		}

		schema.Types[name] = t
	}

	for path, rawEndpoint := range root.Endpoints {
		e := &Endpoint{
			Name:    rawEndpoint.Name,
			Path:    path,
			Args:    make(map[string]Field, len(rawEndpoint.Response.Body)),
			Returns: rawEndpoint.Response.Body,
		}

		if e.Name == "" {
			panic(fmt.Sprintf("`name` is not defined for %s", path))
		}

		normalizedType := strings.TrimPrefix(e.Returns, "[]")
		if _, ok := root.Types[normalizedType]; !ok {
			panic(fmt.Sprintf("Type %s is not defined for %s", normalizedType, path))
		}

		if e.Returns == "" {
			panic(fmt.Sprintf("`returns` is not defined for %s", path))
		}

		schema.Endpoints[path] = e
	}

	return schema, nil
}
