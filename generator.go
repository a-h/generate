package generate

import (
	"bytes"
	"sort"
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

// CreateStructs creates structs from the JSON schema, keyed by the golang name.
func (g *Generator) CreateStructs() map[string]Struct {
	structs := map[string]Struct{}

	// Extract nested and complex types from the JSON schema.
	types := g.schema.ExtractTypes()

	for _, k := range getOrderedKeyNamesFromSchemaMap(types) {
		v := types[k]

		s := Struct{
			ID:     k,
			Name:   getStructName(k, 1),
			Fields: getFields(v.Properties, types),
		}

		structs[s.Name] = s
	}

	return structs
}

func getOrderedKeyNamesFromSchemaMap(m map[string]*jsonschema.Schema) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	sort.Strings(keys)
	return keys
}

func getFields(properties map[string]*jsonschema.Schema, types map[string]*jsonschema.Schema) map[string]Field {
	fields := map[string]Field{}

	for _, k := range getOrderedKeyNamesFromSchemaMap(properties) {
		v := properties[k]

		f := Field{
			Name:     getGolangName(k),
			JSONName: k,
			// Look up the types, try references first, then drop to the built-in types.
			Type: getType(getGolangName(k), v, types),
		}

		fields[f.Name] = f
	}

	return fields
}

func getType(fieldName string, fieldSchema *jsonschema.Schema, types map[string]*jsonschema.Schema) string {
	if _, ok := types[fieldSchema.Reference]; ok {
		return getStructName(fieldSchema.Reference, 1)
	}

	// In the case that the field has properties, then its a complex type and will have a struct
	// generated for it.
	if len(fieldSchema.Properties) > 0 {
		return getGolangName(fieldName)
	}

	return getPrimitiveTypeName(fieldSchema.Type)
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
// The parts refers to the number of segments from the end to take as the name.
func getStructName(reference string, n int) string {
	clean := strings.Replace(reference, "#/", "", -1)
	parts := strings.Split(clean, "/")
	partsToUse := parts[len(parts)-n:]

	sb := bytes.Buffer{}

	for _, p := range partsToUse {
		sb.WriteString(getGolangName(p))
	}

	result := sb.String()

	if result == "" {
		return "Root"
	}

	return result
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
