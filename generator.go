package generate

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"errors"

	"github.com/a-h/generate/jsonschema"
)

// Generator will produce structs from the JSON schema.
type Generator struct {
	schema *jsonschema.Schema
}

// New creates an instance of a generator which will produce structs.
func New(schema *jsonschema.Schema) *Generator {
	return &Generator{
		schema: schema,
	}
}

// CreateStructs creates structs from the JSON schema, keyed by the golang name.
func (g *Generator) CreateStructs() (structs map[string]Struct, err error) {
	structs = make(map[string]Struct)

	// Extract nested and complex types from the JSON schema.
	types := g.schema.ExtractTypes()

	errs := []error{}

	for _, k := range getOrderedKeyNamesFromSchemaMap(types) {
		v := types[k]

		fields, err := getFields(v.Properties, types)

		if err != nil {
			errs = append(errs, err)
		}

		s := Struct{
			ID:     k,
			Name:   getStructName(k, 1),
			Fields: fields,
		}

		structs[s.Name] = s
	}

	if len(errs) > 0 {
		return structs, errors.New(joinErrors(errs))
	}

	return structs, nil
}

func joinErrors(errs []error) string {
	var buffer bytes.Buffer

	for idx, err := range errs {
		buffer.WriteString(err.Error())

		if idx+1 < len(errs) {
			buffer.WriteString(", ")
		}
	}

	return buffer.String()
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

func getFields(properties map[string]*jsonschema.Schema, types map[string]*jsonschema.Schema) (field map[string]Field, err error) {
	fields := map[string]Field{}

	missingTypes := []string{}

	for _, k := range getOrderedKeyNamesFromSchemaMap(properties) {
		v := properties[k]

		tn, typeFound := getType(getGolangName(k), v, types)

		if !typeFound {
			missingTypes = append(missingTypes, getGolangName(k))
		}

		f := Field{
			Name:     getGolangName(k),
			JSONName: k,
			// Look up the types, try references first, then drop to the built-in types.
			Type: tn,
		}

		fields[f.Name] = f
	}

	if len(missingTypes) > 0 {
		return fields, fmt.Errorf("missing types for %s. ", strings.Join(missingTypes, ","))
	}

	return fields, nil
}

func getType(fieldName string, fieldSchema *jsonschema.Schema, types map[string]*jsonschema.Schema) (typeName string, ok bool) {
	if _, ok := types[fieldSchema.Reference]; ok {
		return "*" + getStructName(fieldSchema.Reference, 1), true
	}

	// In the case that the field has properties, then its a complex type and will have a struct
	// generated for it.
	if len(fieldSchema.Properties) > 0 {
		// The '*' is required because the field needs be a pointer to the type to be omitted when nil.
		return "*" + getGolangName(fieldName), true
	}

	// If the type is an object or array, what is it
	// an array of?
	subType := "interface{}"

	// The items property lets us know.
	if fieldSchema.Items != nil {
		// There's a few choices.
		// If there are properties, then a struct has been extracted and title used
		// as the name.
		if len(fieldSchema.Items.Properties) > 0 && fieldSchema.Items.Title != "" {
			subType = fieldSchema.Items.Title
		} else {
			// If that's not set, use the Type property, because it's just a string, number etc.
			if fieldSchema.Items.Type != "" {
				subType = fieldSchema.Items.Type
			}
		}
	}

	if fieldSchema.Reference != "" {
		subType = getStructName(fieldSchema.Reference, 1)
	}

	return getPrimitiveTypeName(fieldSchema.Type, subType)
}

func getPrimitiveTypeName(schemaType string, subType string) (name string, ok bool) {
	switch schemaType {
	case "array":
		return "[]" + subType, true
	case "boolean":
		return "bool", true
	case "integer":
		return "int", true
	case "number":
		return "float64", true
	case "null":
		return "nil", true
	case "object":
		return "*" + subType, true
	case "string":
		return "string", true
	}

	return "undefined", false
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
	stripped := removeAll(s, "_", " ")

	return capitaliseFirstLetter(stripped)
}

func removeAll(s string, remove ...string) string {
	for _, r := range remove {
		s = strings.Replace(s, r, "", -1)
	}
	return s
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
