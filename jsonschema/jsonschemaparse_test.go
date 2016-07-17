package jsonschema

import (
	"strings"
	"testing"
)

func TestThatAMissingSchemaKeyResultsInAnError(t *testing.T) {
	invalid := `{
        "title": "root"
    }`

	_, invaliderr := Parse(invalid)

	valid := `{
        "$schema": "http://json-schema.org/schema#",
        "title": "root"
    }`

	_, validerr := Parse(valid)

	if invaliderr == nil {
		t.Error("When the $schema key is missing from the root, the JSON Schema is not valid")
	}

	if validerr != nil {
		t.Error("It should be possible to parse a simple JSON schema if the $schema key is present")
	}
}

func TestThatTheRootSchemaCanBeParsed(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/schema#",
        "title": "root"
    }`
	so, err := Parse(s)

	if err != nil {
		t.Error("It should be possible to deserialize a simple schema, but recived error ", err)
	}

	if so.Title != "root" {
		t.Errorf("The title was not deserialised from the JSON schema, expected %s, but got %s", "root", so.Title)
	}
}

func TestThatPropertiesCanBeParsed(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/schema#",
        "title": "root",
        "properties": {
            "name": {
                "type": "string"
            },
            "address": {
                "$ref": "#/definitions/address"
            },
            "status": {
                "$ref": "#/definitions/status"
            }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Error("It was not possible to deserialize the schema with references with error ", err)
	}

	tests := []struct {
		actual   func() string
		expected string
	}{
		{
			actual:   func() string { return so.Properties["name"].Type },
			expected: "string",
		},
		{
			actual:   func() string { return so.Properties["address"].Type },
			expected: "",
		},
		{
			actual:   func() string { return so.Properties["address"].Reference },
			expected: "#/definitions/address",
		},
		{
			actual:   func() string { return so.Properties["status"].Reference },
			expected: "#/definitions/status",
		},
	}

	for idx, test := range tests {
		if test.actual() != test.expected {
			t.Errorf("Expected %s but got %s for test %d", test.expected, test.actual(), idx)
		}
	}
}

func TestThatTypesCanBeExtracted(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/draft-04/schema#",
        "id": "Example",
        "definitions": {
            "address": {
                "properties": {
                    "houseName": { "type": "string" },
                    "postcode": { "type": "string" }
                }
            },
            "status": {
                "properties": {
                    "favouritecat": {
                        "enum": [ "A", "B", "C", "D", "E", "F" ],
                        "type": "string"
                    }
                }
            }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	// Check that the definitions have been deserialized correctly into a map.
	if len(so.Definitions) != 2 {
		t.Errorf("The parsed schema should have two child definitions, one for address, one for status, but got %s",
			strings.Join(getKeyNames(so.Definitions), ", "))
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	if len(types) != 3 {
		t.Errorf("expected 3 types, the example, address and status, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the names of the types map to expected references.
	if _, ok := types["#/definitions/address"]; !ok {
		t.Errorf("was expecting to find the address type in the map under key #/definitions/address, available types were %s",
			strings.Join(getKeyNames(types), ", "))
	}
}

func getKeyNames(m map[string]*Schema) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
