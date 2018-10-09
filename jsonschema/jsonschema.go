package jsonschema

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
	Properties           map[string]*Schema
	Required             []string

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

	Examples []string

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

}


// UnmarshalJSON handles unmarshalling AdditionalProperties from JSON.
func (ap *AdditionalProperties) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*ap = (AdditionalProperties)(Schema { AdditionalPropertiesBool: &b })
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
			if k == "oneOf" || k == "allOf" || k == "anyOf" {

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

// returns "type" as an array
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

func (schema *Schema) GetRoot() *Schema {
	if schema.Parent != nil {
		return schema.Parent.GetRoot()
	} else {
		return schema
	}
}

func (schema *Schema) resolveURL(url *url.URL) (*url.URL, error) {
	u, err := url.Parse(schema.GetRoot().ID())
	if err != nil {
		return nil, err
	}

	return u.ResolveReference(url), nil
}

func (schema *Schema) GetDefinitionURL(key string) (*url.URL, error) {
	keyPath, err := url.Parse("#/definitions/"+key)
	if err != nil {
		return nil, err
	}
	return schema.resolveURL(keyPath)
}

func (schema *Schema) ResolveReference() (*url.URL, error) {
	refUrl, err := url.Parse(schema.Reference)
	if err != nil {
		return nil, err
	}
	return schema.resolveURL(refUrl)
}

// Parse parses a JSON schema from a string.
func Parse(schema string) (*Schema, error) {
	s := &Schema{}
	err := json.Unmarshal([]byte(schema), s)

	if err != nil {
		return s, err
	}

	if s.SchemaType == "" {
		return s, errors.New("JSON schema must have a $schema key")
	}

	s.updateParentLinks()

	return s, err
}

