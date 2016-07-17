package generate

import (
	"strings"

	"github.com/a-h/generate/jsonschema"
)

// Generator will produce structs from the JSON schema.
type Generator struct {
	schema *jsonschema.Root
}

// New creates an instance of a generator which will produce structs.
func New(schema *jsonschema.Root) *Generator {
	return &Generator{
		schema: schema,
	}
}

// CreateStructs creates structs from the JSON schema.
func (g *Generator) CreateStructs() []Struct {
	structs := []Struct{}

	types := g.schema.ExtractTypes()

	for k, v := range types {
		s := Struct{
			ID:     k,
			Name:   getStructName(k),
			Fields: getFields(v.Properties, types),
		}

		structs = append(structs, s)
	}

	return structs
}

func getFields(properties map[string]*jsonschema.Schema, types map[string]*jsonschema.Schema) map[string]Field {
	fields := map[string]Field{}

	for k, v := range properties {
		f := Field{
			Name:     getGolangName(k),
			JSONName: k,
			// Look up the types, try references first, then drop to the built-in types.
			Type: getType(v, types),
		}

		fields[f.Name] = f
	}

	return fields
}

func getType(schema *jsonschema.Schema, types map[string]*jsonschema.Schema) string {
	if _, ok := types[schema.Reference]; ok {
		return getStructName(schema.Reference)
	}

	return getPrimitiveTypeName(schema.Type)
}

func getPrimitiveTypeName(schemaType string) string {
	switch schemaType {
	case "array":
		return "[]interface{}"
	case "boolean":
		return "bool"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "null":
		return "nil"
	case "object":
		return "interface{}"
	case "string":
		return "string"
	}

	return "undefined"
}

// getStructName makes a golang struct name from an input reference in the form of #/definitions/address
func getStructName(reference string) string {
	n := strings.Replace(reference, "#/definitions/", "", -1)
	n = strings.Replace(n, "#/", "", -1)

	return getGolangName(n)
}

// getGolangName strips invalid characters out of golang struct or field names.
func getGolangName(s string) string {
	stripped := strings.Replace(s, "_", "", -1)

	return capitaliseFirstLetter(stripped)
}

func capitaliseFirstLetter(s string) string {
	if s == "" {
		return s
	}

	prefix := s[0:1]
	suffix := s[1:]
	return strings.ToUpper(prefix) + suffix
}

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	// The ID within the JSON schema, e.g. #/definitions/address
	ID string
	// The golang name, e.g. "Address"
	Name   string
	Fields map[string]Field
}

// Field defines the data required to generate a field in Go.
type Field struct {
	// The golang name, e.g. "Address1"
	Name string
	// The JSON name, e.g. "address1"
	JSONName string
	// The golang type of the field, e.g. a built-in type like "string" or the name of a struct generated from the JSON schema.
	Type string
}
