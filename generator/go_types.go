package generator

import (
	"fmt"
	"strings"

	"github.com/blakewilliams/overtime/internal/parser"
)

type Endpoint struct {
	endpoint *parser.Endpoint
	schema   *parser.Schema
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

func (ce *Endpoint) Comment() string {
	return formatComment(ce.endpoint.DocComment)
}

func (ce *Endpoint) ResolverMethod() string {
	goType := GoType{parserType: ce.schema.Types[rootType(ce.endpoint.Returns)]}
	if !goType.NeedsResolver() {
		return ""
	}

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

func (gt *GoType) IDType() string {
	return gt.parserType.Fields["id"].Type
}

func (gt *GoType) Comment() string {
	return formatComment(gt.parserType.DocComment)
}

func (gt *GoType) NeedsResolver() bool {
	for _, field := range gt.Fields() {
		if !builtins[field.normalizedType()] {
			return true
		}
	}

	return false
}

func (gt *GoType) Resolvers() []GoResolver {
	resolvers := make([]GoResolver, 0)
	for _, field := range gt.Fields() {
		if builtins[field.normalizedType()] {
			continue
		}

		resolvers = append(resolvers, GoResolver{goType: gt, field: &field})
	}

	return resolvers
}

type GoResolver struct {
	goType *GoType
	field  *GoField
}

func (gr *GoResolver) MethodName() string {
	return fmt.Sprintf(
		"Resolve%s%s",
		gr.goType.Name(),
		gr.field.Name(),
	)
}

func (gr *GoResolver) Arguments() string {
	return fmt.Sprintf(
		"%sIDs []%s",
		uncapitalize(gr.goType.Name()),
		gr.goType.IDType(),
	)
}

func (gr *GoResolver) Comment() string {
	return formatComment(fmt.Sprintf(`Populates the %s field for the %s type`, gr.field.Name(), gr.goType.Name()))
}

func (gr *GoResolver) ReturnType() string {
	if strings.HasPrefix(gr.field.Type(), "[]") {
		return "map[int64][]" + capitalize(strings.TrimPrefix(gr.field.Type(), "[]"))
	} else {
		return "map[int64]" + capitalize(gr.field.Type())
	}
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

func (gf *GoField) Comment() string {
	return formatComment(gf.parserField.DocComment)
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

func (gf *GoField) IsBuiltin() bool {
	return builtins[gf.normalizedType()]
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

func rootType(t string) string {
	return strings.TrimPrefix(t, "[]")
}
