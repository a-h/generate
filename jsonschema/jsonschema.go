package jsonschema

import (
	"encoding/json"
	"errors"
)

// Schema represents JSON schema.
type Schema struct {
	SchemaType  string `json:"$schema"`
	Title       string `json:"title"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Definitions map[string]*Schema
	Properties  map[string]*Schema
	Reference   string `json:"$ref"`
	// Items represents the types that are permitted in the array.
	Items *Schema `json:"items"`
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

	return s, err
}

// ExtractTypes creates a map of defined types within the schema.
func (s *Schema) ExtractTypes() map[string]*Schema {
	types := make(map[string]*Schema)

	// Pass in the # to start the path off.
	id := s.ID

	// If there's no ID, try the title instead. Otherwise, the struct will get the name
	// of "Root".
	if id == "" {
		id = s.Title
	}

	addTypeAndChildrenToMap("#", id, s, types)

	return types
}

func addTypeAndChildrenToMap(path string, name string, s *Schema, types map[string]*Schema) {
	if s.Type == "array" {
		addTypeAndChildrenToMap(path, s.Items.Title, s.Items, types)
	}

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

// ListReferences lists all of the references in a schema.
func (s *Schema) ListReferences() map[string]bool {
	m := make(map[string]bool)
	addReferencesToMap(s, m)
	return m
}

func addReferencesToMap(s *Schema, m map[string]bool) {
	if s.Reference != "" {
		m[s.Reference] = true
	}

	if s.Definitions != nil {
		for _, d := range s.Definitions {
			addReferencesToMap(d, m)
		}
	}

	if s.Properties != nil {
		for _, p := range s.Properties {
			addReferencesToMap(p, m)
		}
	}

	if s.Items != nil {
		addReferencesToMap(s.Items, m)
	}
}
