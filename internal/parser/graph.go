package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type (
	// Graph represents the entire schema for all federated services and is
	// responsible for combining the relevant schemas of each service into a
	// single schema that can be used to generate a gateway.
	Graph struct {
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

// APIName returns the user and API friendly name for the endpoint. It uses the path parameters to
// generate the name using conventions.
//
// For example, the endpoint "/api/v1/users/:user_id/comments" would return "ApiV1UserComments"
//
// The User is singular instead of plural because it has a path parameter that does not end in an s
func (e *Endpoint) APIName() string {
	parts := strings.Split(e.Path, "/")
	nameFromPath := strings.Builder{}

	for i, part := range parts {
		if part == "" {
			continue
		}

		if len(parts) > i+1 && strings.HasPrefix(parts[i+1], ":") {
			continue
		}

		if strings.HasPrefix(part, ":") {
			nameFromPath.WriteString("By" + capitalize(part[1:]))
		} else {
			nameFromPath.WriteString(capitalize(part))
		}
	}

	return nameFromPath.String()
}

func isSingular(s string) bool {
	word := strings.ToLower(s)

	if strings.HasSuffix(word, "ss") {
		return true
	}

	// Step 2: Check for common singular endings
	if strings.HasSuffix(word, "us") || strings.HasSuffix(word, "is") {
		return true
	}

	// Step 3: Check for common plural endings
	if strings.HasSuffix(word, "ies") || strings.HasSuffix(word, "es") {
		return false
	}

	if strings.HasSuffix(word, "id") {
		return true
	}

	if strings.HasSuffix(word, "ids") {
		return false
	}

	// Step 4: Check for "s" with a preceding vowel or consonant
	if strings.HasSuffix(word, "s") {
		if len(word) > 1 {
			// Get the second last character
			secondLast := word[len(word)-2]
			// Check if the second last character is a vowel
			if strings.ContainsRune("aeiou", rune(secondLast)) {
				return true // Likely singular
			} else {
				return false // Likely plural
			}
		}
	}

	// If none of the above conditions are met, assume singular
	return true
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)

	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}
