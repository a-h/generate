package jsonschema

import (
	"fmt"
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
			"arrayOfTypes": {
                "type": ["string", "nonsense"]
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
			actual:   func() string { return so.Properties["name"].Type()[0] },
			expected: "string",
		},
		{
			actual:   func() string { return so.Properties["arrayOfTypes"].Type()[0] },
			expected: "string",
		},
		{
			actual:   func() string { return so.Properties["arrayOfTypes"].Type()[1] },
			expected: "nonsense",
		},
		{
			actual:   func() string { return so.Properties["address"].Type()[0] },
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
            },
			"links": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"asset_id": {
							"type": "string"
						}
					}
				}
			}
        },
        "properties": {
            "address": { "$ref": "#/definitions/address" }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	// Check that the definitions have been deserialized correctly into a map.
	if len(so.Definitions) != 3 {
		t.Errorf("The parsed schema should have two child definitions, one for address, one for status, and one for links but got %s",
			strings.Join(getKeyNames(so.Definitions), ", "))
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	if len(types) != 4 {
		t.Errorf("expected 4 types, the example, address, status and links, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the names of the types map to expected references.
	if _, ok := types["#/definitions/address"]; !ok {
		t.Errorf("was expecting to find the address type in the map under key #/definitions/address, available types were %s",
			strings.Join(getKeyNames(types), ", "))
	}

	if _, ok := types["#/definitions/links"]; !ok {
		t.Errorf("was expecting to find the links type in the map under key #/definitions/links, available types were %s",
			strings.Join(getKeyNames(types), ", "))
	}
}

func TestThatNestedTypesCanBeExtracted(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/draft-04/schema#",
        "id": "Example",
        "properties": {
            "favouritecat": {
                "enum": [ "A", "B", "C", "D", "E", "F" ],
                "type": "string"
            },
            "subobject": {
                "type": "object",
                "properties": {
                    "subproperty1": {
                        "type": "date"
                    }
                }
            }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	if len(types) != 2 {
		t.Errorf("expected 2 types, the example and subobject, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the names of the types map to expected references.
	if _, ok := types["#/properties/subobject"]; !ok {
		t.Errorf("was expecting to find the subobject type in the map under key #/properties/subobject, available types were %s",
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

func TestThatArraysAreSupported(t *testing.T) {
	s := `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "title": "ProductSet",
    "type": "array",
    "items": {
        "title": "Product",
        "type": "object",
        "properties": {
            "id": {
                "description": "The unique identifier for a product",
                "type": "number"
            },
            "name": {
                "type": "string"
            },
            "price": {
                "type": "number",
                "minimum": 0,
                "exclusiveMinimum": true
            },
            "tags": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "minItems": 1,
                "uniqueItems": true
            }
        },
        "required": ["id", "name", "price"]
    }
}`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	fmt.Printf("Types: %+v\n", types)

	if len(types) != 1 {
		t.Errorf("expected 1 type, just the Product, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the names of the types map to expected references.
	ps, ok := types["#/Product"]
	if !ok {
		t.Fatalf("was expecting to find the Product type but available types were %s",
			strings.Join(getKeyNames(types), ", "))
	}

	if len(ps.Properties) != 4 {
		t.Errorf("was expecting the Product to have 4 properties, but it had %d", len(ps.Properties))
	}

	if !ps.Properties["tags"].IsArray() {
		t.Errorf("expected the 'Tags' property type to be an array, but it was %s", ps.Properties["tags"].Type())
	}
}

func TestThatReferencesCanBeListed(t *testing.T) {
	s := `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "title": "Product set",
    "type": "array",
    "items": {
        "title": "Product",
        "type": "object",
        "properties": {
            "id": {
                "description": "The unique identifier for a product",
                "type": "number"
            },
            "name": {
                "type": "string"
            },
            "price": {
                "type": "number",
                "minimum": 0,
                "exclusiveMinimum": true
            },
            "tags": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "minItems": 1,
                "uniqueItems": true
            },
            "dimensions": {
      			"$ref": "#/definitions/address"
            },
            "warehouseLocation": {
                "description": "Coordinates of the warehouse with the product",
                "$ref": "http://json-schema.org/geo"
            }
        },
        "required": ["id", "name", "price"]
    }
}`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	refs := so.ListReferences()

	if len(refs) != 2 {
		t.Errorf("Expected 1 references, one internal, one external, but got %d references", len(refs))
	}

	if _, ok := refs["http://json-schema.org/geo"]; !ok {
		t.Error("Couldn't find the reference to http://json-schema.org/geo")
	}

	if _, ok := refs["#/definitions/address"]; !ok {
		t.Error("Couldn't find the reference to #/definitions/address")
	}
}

func TestThatRequiredPropertiesAreIncludedInTheSchemaModel(t *testing.T) {
	s := `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "name": "Repository Configuration",
    "type": "object",
    "additionalProperties": false,
    "required": [ "name" ],
    "properties": {
        "name": {
            "type": "string",
            "description": "Repository name."
        },
        "repositories": {
            "type": "string",
            "description": "A set of additional repositories where packages can be found.",
            "additionalProperties": true
        }
    }
}`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	types := so.ExtractTypes()

	if len(types) != 1 {
		t.Errorf("Expected just the Repository Configuration type to be extracted, but got %d types extracted", len(types))
	}

	var rc *Schema
	var ok bool
	if rc, ok = types["#"]; !ok {
		t.Fatalf("Couldn't find the reference to the Repository Configuration root type, the types found were %+v", getKeyNames(types))
	}

	if len(rc.Required) != 1 || rc.Required[0] != "name" {
		t.Errorf("Expected the required field of the Repository Configuration type to contain a reference to 'name'.")
	}
}
