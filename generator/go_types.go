package generator

import (
	"fmt"
	"strings"

	"github.com/blakewilliams/overtime/internal/parser"
)

type Endpoint struct {
	endpoint *parser.Endpoint
}

func (ce *Endpoint) Method() string {
	return ce.endpoint.Method
}

func (ce *Endpoint) MethodName() string {
	return capitalize(ce.endpoint.Name)
}

func (ce *Endpoint) ReturnValue() string {
	if strings.HasPrefix(ce.endpoint.Returns, "[]") {
		return "[]" + "*" + strings.TrimPrefix(ce.endpoint.Returns, "[]")
	} else {
		return "*" + ce.endpoint.Returns
	}
}

func (ce *Endpoint) Path() string {
	parts := strings.Split(ce.endpoint.Path, "/")
	formattedParts := make([]string, len(parts))

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			formattedParts[i] = fmt.Sprintf("{%s}", strings.TrimPrefix(part, ":"))
			continue
		}

		formattedParts[i] = part
	}
	return strings.Join(formattedParts, "/")
}

func (ce *Endpoint) ResolverMethod() string {
	if strings.HasPrefix(ce.endpoint.Returns, "[]") {
		return fmt.Sprintf("ResolveFor%s(result, c.resolver)", capitalize(strings.TrimPrefix(ce.endpoint.Returns, "[]")))
	}

	return fmt.Sprintf(
		"ResolveFor%s([]*%s{result}, c.resolver)",
		capitalize(strings.TrimPrefix(ce.endpoint.Returns, "[]")),
		capitalize(strings.TrimPrefix(ce.endpoint.Returns, "[]")),
	)
}

type GoType struct {
	parserType *parser.Type
}

func (gt *GoType) Name() string {
	if gt.parserType.Name == "id" {
		return "ID"
	}

	return capitalize(gt.parserType.Name)
}

func (gt *GoType) MapType() string {
	if strings.HasPrefix(gt.Name(), "[]") {
		return "map[int64][]*" + capitalize(strings.TrimPrefix(gt.Name(), "[]"))
	} else {
		return "map[int64]*" + capitalize(gt.Name())
	}
}

func (gt *GoType) Fields() []GoField {
	fields := make([]GoField, 0, len(gt.parserType.Fields))

	for _, field := range gt.parserType.Fields {
		fields = append(fields, GoField{parserField: field, parentType: gt})
	}

	return fields
}

type GoField struct {
	parserField parser.Field
	parentType  *GoType
}

func (gf *GoField) Name() string {
	if gf.parserField.Name == "id" {
		return "ID"
	}

	return capitalize(gf.parserField.Name)
}

func (gf *GoField) Type() string {
	if builtins[gf.normalizedType()] {
		return gf.parserField.Type
	}
	if strings.HasPrefix(gf.parserField.Type, "[]") && !builtins[gf.normalizedType()] {
		return "[]" + "*" + strings.TrimPrefix(gf.parserField.Type, "[]")
	} else {
		return "*" + gf.parserField.Type
	}
}

func (gf *GoField) normalizedType() string {
	return strings.TrimPrefix(gf.parserField.Type, "[]")
}

func (gf *GoField) Tags() string {
	tag := strings.Builder{}
	tag.WriteString(" `")
	tag.Write([]byte(fmt.Sprintf("json:\"%s\"", gf.parserField.Name)))
	if !builtins[gf.normalizedType()] {
		tag.Write([]byte(fmt.Sprintf(" resolver:\"%s\"", gf.ResolverMethodName())))
		// TODO backfill
		// resolvers[resolverName] = field
	}
	tag.WriteString("`")

	return tag.String()
}

func (gf *GoField) ResolverMethodName() string {
	return fmt.Sprintf(
		"Resolve%s%s",
		gf.parentType.Name(),
		gf.Name(),
	)
}
