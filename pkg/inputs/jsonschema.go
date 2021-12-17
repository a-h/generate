package inputs

import (
	"encoding/json"
	"errors"
	"net/url"
)

// AdditionalProperties handles additional properties present in the JSON schema.
type AdditionalProperties Schema

// Schema represents JSON schema.
type Schema struct {
	// SchemaType identifies the schema version.
	// http://json-schema.org/draft-07/json-schema-core.html#rfc.section.7
	SchemaType string `json:"$schema"`

	// ID{04,06} is the schema URI identifier.
	// http://json-schema.org/draft-07/json-schema-core.html#rfc.section.8.2
	ID04 string `json:"id"`  // up to draft-04
	ID06 string `json:"$id"` // from draft-06 onwards

	// Title and Description state the intent of the schema.
	Title       string
	Description string

	// TypeValue is the schema instance type.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.6.1.1
	TypeValue interface{} `json:"type"`

	// Definitions are inline re-usable schemas.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.9
	Definitions map[string]*Schema

	// Properties, Required and AdditionalProperties describe an object's child instances.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.6.5
	Properties map[string]*Schema
	Required   []string

	// "additionalProperties": {...}
	AdditionalProperties *AdditionalProperties

	// "additionalProperties": false
	AdditionalPropertiesBool *bool `json:"-"`

	AnyOf []*Schema
	AllOf []*Schema
	OneOf []*Schema

	// Default can be used to supply a default JSON value associated with a particular schema.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.10.2
	Default interface{}

	// Examples ...
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.10.4
	Examples []interface{}

	// Reference is a URI reference to a schema.
	// http://json-schema.org/draft-07/json-schema-core.html#rfc.section.8
	Reference string `json:"$ref"`

	// Items represents the types that are permitted in the array.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.6.4
	Items *Schema

	// NameCount is the number of times the instance name was encountered across the schema.
	NameCount int `json:"-" `

	// Parent schema
	Parent *Schema `json:"-" `

	// Key of this schema i.e. { "JSONKey": { "type": "object", ....
	JSONKey string `json:"-" `

	// path element - for creating a path by traversing back to the root element
	PathElement string `json:"-"`

	// calculated struct name of this object, cached here
	GeneratedType string `json:"-"`
}

// UnmarshalJSON handles unmarshalling AdditionalProperties from JSON.
func (ap *AdditionalProperties) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*ap = (AdditionalProperties)(Schema{AdditionalPropertiesBool: &b})
		return nil
	}

	// support anyOf, allOf, oneOf
	a := map[string][]*Schema{}
	if err := json.Unmarshal(data, &a); err == nil {
		for k, v := range a {
			switch k {
			case "oneOf":
				ap.OneOf = append(ap.OneOf, v...)
			case "allOf":
				ap.AllOf = append(ap.AllOf, v...)
			case "anyOf":
				ap.AnyOf = append(ap.AnyOf, v...)
			}
		}
		return nil
	}

	s := Schema{}
	err := json.Unmarshal(data, &s)
	if err == nil {
		*ap = AdditionalProperties(s)
	}
	return err
}

// ID returns the schema URI id.
func (schema *Schema) ID() string {
	// prefer "$id" over "id"
	if schema.ID06 == "" && schema.ID04 != "" {
		return schema.ID04
	}
	return schema.ID06
}

// Type returns the type which is permitted or an empty string if the type field is missing.
// The 'type' field in JSON schema also allows for a single string value or an array of strings.
// Examples:
//   "a" => "a", false
//   [] => "", false
//   ["a"] => "a", false
//   ["a", "b"] => "a", true
func (schema *Schema) Type() (firstOrDefault string, multiple bool) {
	// We've got a single value, e.g. { "type": "object" }
	if ts, ok := schema.TypeValue.(string); ok {
		firstOrDefault = ts
		multiple = false
		return
	}

	// We could have multiple types in the type value, e.g. { "type": [ "object", "array" ] }
	if a, ok := schema.TypeValue.([]interface{}); ok {
		multiple = len(a) > 1
		for _, n := range a {
			if s, ok := n.(string); ok {
				firstOrDefault = s
				return
			}
		}
	}

	return "", multiple
}

// MultiType returns "type" as an array
func (schema *Schema) MultiType() ([]string, bool) {
	// We've got a single value, e.g. { "type": "object" }
	if ts, ok := schema.TypeValue.(string); ok {
		return []string{ts}, false
	}

	// We could have multiple types in the type value, e.g. { "type": [ "object", "array" ] }
	if a, ok := schema.TypeValue.([]interface{}); ok {
		rv := []string{}
		for _, n := range a {
			if s, ok := n.(string); ok {
				rv = append(rv, s)
			}
		}
		return rv, len(rv) > 1
	}

	return nil, false
}

// GetRoot returns the root schema.
func (schema *Schema) GetRoot() *Schema {
	if schema.Parent != nil {
		return schema.Parent.GetRoot()
	}
	return schema
}

// Parse parses a JSON schema from a string.
func Parse(schema string, uri *url.URL) (*Schema, error) {
	return ParseWithSchemaKeyRequired(schema, uri, true)
}

// ParseWithSchemaKeyRequired parses a JSON schema from a string with a flag to set whether the schema key is required.
func ParseWithSchemaKeyRequired(schema string, uri *url.URL, schemaKeyRequired bool) (*Schema, error) {
	s := &Schema{}
	err := json.Unmarshal([]byte(schema), s)

	if err != nil {
		return s, err
	}

	if s.ID() == "" {
		s.ID06 = uri.String()
	}

	if schemaKeyRequired && s.SchemaType == "" {
		return s, errors.New("JSON schema must have a $schema key unless schemaKeyRequired flag is set")
	}

	// validate root URI, it MUST be an absolute URI
	abs, err := url.Parse(s.ID())
	if err != nil {
		return nil, errors.New("error parsing $id of document \"" + uri.String() + "\": " + err.Error())
	}
	if !abs.IsAbs() {
		return nil, errors.New("$id of document not absolute URI: \"" + uri.String() + "\": \"" + s.ID() + "\"")
	}

	s.Init()

	return s, nil
}

// Init schema.
func (schema *Schema) Init() {
	root := schema.GetRoot()
	root.updateParentLinks()
	root.ensureSchemaKeyword()
	root.updatePathElements()
}

func (schema *Schema) updatePathElements() {
	if schema.IsRoot() {
		schema.PathElement = "#"
	}

	for k, d := range schema.Definitions {
		d.PathElement = "definitions/" + k
		d.updatePathElements()
	}

	for k, p := range schema.Properties {
		p.PathElement = "properties/" + k
		p.updatePathElements()
	}

	if schema.AdditionalProperties != nil {
		schema.AdditionalProperties.PathElement = "additionalProperties"
		(*Schema)(schema.AdditionalProperties).updatePathElements()
	}

	if schema.Items != nil {
		schema.Items.PathElement = "items"
		schema.Items.updatePathElements()
	}
}

func (schema *Schema) updateParentLinks() {
	for k, d := range schema.Definitions {
		d.JSONKey = k
		d.Parent = schema
		d.updateParentLinks()
	}

	for k, p := range schema.Properties {
		p.JSONKey = k
		p.Parent = schema
		p.updateParentLinks()
	}
	if schema.AdditionalProperties != nil {
		schema.AdditionalProperties.Parent = schema
		(*Schema)(schema.AdditionalProperties).updateParentLinks()
	}
	if schema.Items != nil {
		schema.Items.Parent = schema
		schema.Items.updateParentLinks()
	}
}

func (schema *Schema) ensureSchemaKeyword() error {
	check := func(k string, s *Schema) error {
		if s.SchemaType != "" {
			return errors.New("invalid $schema keyword: " + k)
		}
		return s.ensureSchemaKeyword()
	}
	for k, d := range schema.Definitions {
		if err := check(k, d); err != nil {
			return err
		}
	}
	for k, d := range schema.Properties {
		if err := check(k, d); err != nil {
			return err
		}
	}
	if schema.AdditionalProperties != nil {
		if err := check("additionalProperties", (*Schema)(schema.AdditionalProperties)); err != nil {
			return err
		}
	}
	if schema.Items != nil {
		if err := check("items", schema.Items); err != nil {
			return err
		}
	}
	return nil
}

// FixMissingTypeValue is backwards compatible, guessing the users intention when they didn't specify a type.
func (schema *Schema) FixMissingTypeValue() {
	if schema.TypeValue == nil {
		if schema.Reference == "" && len(schema.Properties) > 0 {
			schema.TypeValue = "object"
			return
		}
		if schema.Items != nil {
			schema.TypeValue = "array"
			return
		}
	}
}

// IsRoot returns true when the schema is the root.
func (schema *Schema) IsRoot() bool {
	return schema.Parent == nil
}
