package generate

import (
	"fmt"
	"bytes"
	"strings"
	"unicode"
	"net/url"
	"errors"
	"sort"
	"path"
	"github.com/a-h/generate/jsonschema"
)

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	// The ID within the JSON schema, e.g. #/definitions/address
	ID string
	// The golang name, e.g. "Address"
	Name string
	// Description of the struct
	Description string
	Fields      map[string]Field
	AdditionalValueType string
}

// Field defines the data required to generate a field in Go.
type Field struct {
	// The golang name, e.g. "Address1"
	Name string
	// The JSON name, e.g. "address1"
	JSONName string
	// The golang type of the field, e.g. a built-in type like "string" or the name of a struct generated
	// from the JSON schema.
	Type string
	// Required is set to true when the field is required.
	Required bool
	Comment  string
}

// Generator will produce structs from the JSON schema.
type Generator struct {
	schemas   []*jsonschema.Schema
	Structs   map[string]Struct
	Aliases   map[string]Field
	// cache for reference types
	refs      map[string]string
	anonCount int
}

// New creates an instance of a generator which will produce structs.
func New(schemas ...*jsonschema.Schema) *Generator {
	return &Generator{
		schemas: schemas,
		Structs: make(map[string]Struct),
		Aliases: make(map[string]Field),
		refs:    make(map[string]string),
	}
}

// CreateTypes creates types from the JSON schemas, keyed by the golang name.
func (g *Generator) CreateTypes() (err error) {
	// process the root node
	for _, schema := range g.schemas {
		name := g.getSchemaName(schema)
		if rootType, err := g.processSchema(name, schema); err != nil {
			return err
		} else {
			if rootType == "interface{}" {
				a := Field {
					Name:     name,
					JSONName: "",
					Type:     rootType,
					Required: false,
					Comment:  schema.Description,
				}
				g.Aliases[a.Name] = a
			}
		}
	}
	return
}

func (g *Generator) processDefinitions(root *jsonschema.Schema) error {
	keys := getOrderedKeyNamesFromSchemaMap(root.Definitions)
	// now do the actual work.
	for _, key := range keys {
		if refUrl, err := root.GetDefinitionURL(key); err != nil {
			return err
		} else {
			if g.refs[refUrl.String()] != "" {
				// already processed by dependency
				continue
			} else {
				if _, err := g.processDefinition(refUrl, root); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *Generator) processDefinition(u *url.URL, schema *jsonschema.Schema) (typ string, err error) {
	root := schema.GetRoot()
	defKey := path.Base(u.String())
	node, ok := root.Definitions[defKey]
	if !ok {
		return "", errors.New("definition not found: " + u.String())
	}
	if typ, err = g.processSchema(getGolangName(defKey), node); err != nil {
		return
	}
	g.refs[u.String()] = typ
	return
}

// returns the type refered to by schema after resolving all dependencies
func (g *Generator) processSchema(key string, schema *jsonschema.Schema) (typ string, err error) {
	if len(schema.Definitions) > 0 {
		g.processDefinitions(schema)
	}
	// if we have multiple schema types, the golang type will be interface{}
	typ = "interface{}"
	types, isMultiType := schema.MultiType()
	if len(types) > 0 {
		for _, schemaType := range types {
			name := key
			if isMultiType {
				name = name + "_" + schemaType
			}
			switch schemaType {
			case "object":
				if rv, err := g.processObject(name, schema); err != nil {
					return "", err
				} else {
					if !isMultiType {
						return rv, nil
					}
				}
			case "array":
				if rv, err := g.processArray(name, schema); err != nil {
					return "", err
				} else {
					if !isMultiType {
						return rv, nil
					}
				}
			default:
				if rv, err := getPrimitiveTypeName(schemaType, "", false); err != nil {
					return "", err
				} else {
					if !isMultiType {
						return rv, nil
					}
				}
			}
		}
	} else {
		if schema.Reference != "" {
			if refUrl, err := schema.ResolveReference(); err != nil {
				return "", err
			} else {
				typ = g.refs[refUrl.String()]
				if typ == "" {
					return g.processDefinition(refUrl, schema)
				}
			}
		}
	}
	return // return interface{}
}

// name: name of this array, usually the js key
// schema: items element
func (g *Generator) processArray(name string, schema *jsonschema.Schema) (typeStr string, err error) {
	if schema.Items != nil {
		subName := g.getSchemaName(schema.Items)
		if subName == "" {
			subName = name + "Items"
		}
		subTyp, err := g.processSchema(subName, schema.Items)
		if err != nil {
			return "", err
		}
		if finalType, err := getPrimitiveTypeName("array", subTyp, true); err != nil {
			return "", err
		} else {
			// only alias root arrays
			if schema.Parent == nil {
				array := Field{
					Name:     name,
					JSONName: "",
					Type:     finalType,
					Required: contains(schema.Required, name),
					Comment:  schema.Description,
				}
				g.Aliases[array.Name] = array
			}
			return finalType, nil
		}
	}
	return "[]interface{}", nil
}

func (g *Generator) processObject(name string, schema *jsonschema.Schema) (typ string, err error) {
	strct := Struct{
		ID:          schema.ID(),
		Name:        name,
		Description: schema.Description,
		Fields:      make(map[string]Field, len(schema.Properties)),
	}
	for propKey, prop := range schema.Properties {
		propName := getGolangName(propKey)
		if subTyp, err := g.processSchema(propName, prop); err != nil {
			return "", err
		} else {
			f := Field{
				Name:     propName,
				JSONName: propKey,
				Type:     subTyp,
				Required: contains(schema.Required, propKey),
				Comment:  prop.Description,
			}
			strct.Fields[f.Name] = f
		}
	}
	// additionalProperties with some kind of typed schema
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.AdditionalPropertiesBool == nil {
		ap := (*jsonschema.Schema)(schema.AdditionalProperties)
		apName := g.getSchemaName(ap)
		subTyp, err := g.processSchema(apName, ap)
		if err != nil {
			return "", err
		}
		// since this struct will have extra fields and of a known type, emit code to parse them...
		strct.AdditionalValueType = subTyp
		// and add a field to the struct to store the additional stuff
		subTyp = "map[string]" + subTyp
		f := Field{
			Name:     "AdditionalProperties",
			JSONName: "-",
			Type:     subTyp,
			Required: false,
			Comment:  "",
		}
		strct.Fields[f.Name] = f
	}
	// additionalProperties as either true (everything) or false (nothing)
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.AdditionalPropertiesBool != nil {
		if *schema.AdditionalProperties.AdditionalPropertiesBool == true {
			// everything
			subTyp := "map[string]interface{}"
			f := Field{
				Name:     "AdditionalProperties",
				JSONName: "-",
				Type:     subTyp,
				Required: false,
				Comment:  "",
			}
			strct.Fields[f.Name] = f
		} else {
			// nothing
		}
	}
	g.Structs[strct.Name] = strct
	// objects are always a pointer
	return getPrimitiveTypeName("object", name, true)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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

func getPrimitiveTypeName(schemaType string, subType string, pointer bool) (name string, err error) {
	switch schemaType {
	case "array":
		if subType == "" {
			return "error_creating_array", errors.New("can't create an array of an empty subtype")
		}
		return "[]" + subType, nil
	case "boolean":
		return "bool", nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "null":
		return "nil", nil
	case "object":
		if subType == "" {
			return "error_creating_object", errors.New("can't create an object of an empty subtype")
		}
		if pointer {
			return "*" + subType, nil
		}
		return subType, nil
	case "string":
		return "string", nil
	}

	return "undefined", fmt.Errorf("failed to get a primitive type for schemaType %s and subtype %s",
		schemaType, subType)
}

// return a name for this (sub-)schema. TODO: move to *Schema receivership
func (g *Generator) getSchemaName(schema *jsonschema.Schema) (string) {
	if len(schema.Title) > 0 {
		return getGolangName(schema.Title)
	}

	if schema.Parent == nil {
		rootName := schema.Title

		if rootName == "" {
			rootName = schema.Description
		}

		if rootName == "" {
			rootName = "Root"
		}

		return getGolangName(rootName)
	}

	if schema.JSONKey != "" {
		return getGolangName(schema.JSONKey)
	}
	if schema.Parent != nil && schema.Parent.JSONKey != "" {
		// ugh...
		return getGolangName(schema.Parent.JSONKey + "Item")
	}

	g.anonCount ++
	return fmt.Sprintf("Anonymous%d", g.anonCount)
}

// getGolangName strips invalid characters out of golang struct or field names.
func getGolangName(s string) string {
	buf := bytes.NewBuffer([]byte{})

	for i, v := range splitOnAll(s, isNotAGoNameCharacter) {
		if i == 0 && strings.IndexAny(v, "0123456789") == 0 {
			// Go types are not allowed to start with a number, lets prefix with an underscore.
			buf.WriteRune('_')
		}
		buf.WriteString(capitaliseFirstLetter(v))
	}

	return buf.String()
}

func splitOnAll(s string, shouldSplit func(r rune) bool) []string {
	rv := []string{}

	buf := bytes.NewBuffer([]byte{})
	for _, c := range s {
		if shouldSplit(c) {
			rv = append(rv, buf.String())
			buf.Reset()
		} else {
			buf.WriteRune(c)
		}
	}
	if buf.Len() > 0 {
		rv = append(rv, buf.String())
	}

	return rv
}

func isNotAGoNameCharacter(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	return true
}

func capitaliseFirstLetter(s string) string {
	if s == "" {
		return s
	}

	prefix := s[0:1]
	suffix := s[1:]
	return strings.ToUpper(prefix) + suffix
}
