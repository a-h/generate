package jsonschema

import (
	"encoding/json"
	"errors"
)

// Root represents the root JSON Schema, which can be used to generate structs.
type Root struct {
	Schema
	SchemaType string `json:"$schema"`
}

// Schema represents JSON schema.
type Schema struct {
	Title       string `json:"title"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Definitions map[string]*Schema
	Properties  map[string]*Schema
	Reference   string `json:"$ref"`
}

// Parse parses a JSON schema from a string.
func Parse(schema string) (*Root, error) {
	s := &Root{}
	err := json.Unmarshal([]byte(schema), s)

	if err != nil {
		return s, err
	}

	if s.SchemaType == "" {
		return s, errors.New("JSON schema must have a $schema key")
	}

	return s, err
}

// ExtractTypes creates a map of defined types within the schema.
func (s *Root) ExtractTypes() map[string]*Schema {
	types := make(map[string]*Schema)

	// Pass in the # to start the path off.
	addTypeAndChildrenToMap("#", s.ID, &s.Schema, types)

	return types
}

func addTypeAndChildrenToMap(path string, name string, s *Schema, types map[string]*Schema) {
	if len(s.Properties) > 0 {
		types[path+"/"+name] = s
	}

	if s.Definitions != nil {
		for k, d := range s.Definitions {
			addTypeAndChildrenToMap(path+"/definitions", k, d, types)
		}
	}

	if s.Properties != nil {
		for k, d := range s.Properties {
			// Only add the children as their own type if they have properties at all.
			addTypeAndChildrenToMap(path+"/"+"properties", k, d, types)
		}
	}
}
