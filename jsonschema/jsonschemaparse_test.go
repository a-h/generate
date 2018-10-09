package jsonschema

import (
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
		t.Fatal("It should be possible to unmarshal a simple schema, but received error:", err)
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
		t.Fatal("It was not possible to unmarshal the schema:", err)
	}

	nameType, nameMultiple := so.Properties["name"].Type()
	if nameType != "string" || nameMultiple {
		t.Errorf("expected property 'name' type to be 'string', but was '%v'", nameType)
	}

	addressType, _ := so.Properties["address"].Type()
	if addressType != "" {
		t.Errorf("expected property 'address' type to be '', but was '%v'", addressType)
	}

	if so.Properties["address"].Reference != "#/definitions/address" {
		t.Errorf("expected property 'address' reference to be '#/definitions/address', but was '%v'", so.Properties["address"].Reference)
	}

	if so.Properties["status"].Reference != "#/definitions/status" {
		t.Errorf("expected property 'status' reference to be '#/definitions/status', but was '%v'", so.Properties["status"].Reference)
	}
}

/*
func TestThatStructTypesCanBeExtracted(t *testing.T) {
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
		t.Fatal("Failed to parse the test JSON: ", err)
	}

	// Check that the definitions have been unmarshalled correctly into a map.
	if len(so.Definitions) != 3 {
		t.Errorf("The parsed schema should have two child definitions, one for address, one for status, and one for links but got %s",
			strings.Join(getKeyNames(so.Definitions), ", "))
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	if len(types) != 4 {
		t.Errorf("Expected 4 types, the example, address, status and links, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the keys of the types map to expected references.
	if _, ok := types["#/definitions/address"]; !ok {
		t.Errorf("Expecting to find the address type in the map under key #/definitions/address, available keys were %s",
			strings.Join(getKeyNames(types), ", "))
	}

	if _, ok := types["#/definitions/links"]; !ok {
		t.Errorf("Expected to find the links type in the map under key #/definitions/links, available keys were %s",
			strings.Join(getKeyNames(types), ", "))
	}
}

func TestThatAliasTypesCanBeExtracted(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/draft-07/schema#",
        "$id": "Example",
        "type": "string",
        "enum": [ "A", "B", "C", "D", "E", "F" ]
    }`
	so, err := Parse(s)

	if err != nil {
		t.Fatal("Failed to parse the test JSON: ", err)
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	if len(types) != 1 {
		t.Errorf("Expected 1 type, the Root type, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the key of the type maps to expected reference.
	if _, ok := types["#"]; !ok {
		t.Errorf("Expected to find the Root type in the map under key #, available keys were %s",
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
		t.Errorf("was expecting to find the subobject type in the map under key #/properties/subobject, available types were '%s'",
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

func TestThatAdditionalPropertiesCanBeExtracted(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/draft-04/schema#",
        "id": "Example",
        "properties": {
            "subobject1": {
                "type": "object",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "subobject2": {
                "type": "object",
                "additionalProperties": {
                    "type": "object",
                    "properties": {
                        "x": { "type": "string" }
                    }
                }
            },
            "subobject3": {
                "type": "object",
                "additionalProperties": {
                    "anyOf": [
                        {
                            "type": "object",
                            "properties": {
                                "x": { "type": "string" }
                            }
                        }
                    ],
                    "allOf": [
                        {
                            "type": "object",
                            "properties": {
                                "x": { "type": "string" }
                            }
                        }
                    ],
                    "oneOf": [
                        {
                            "type": "object",
                            "properties": {
                                "x": { "type": "string" }
                            }
                        }
                    ],
                    "not": [
                        {
                            "type": "object",
                            "properties": {
                                "x": { "type": "string" }
                            }
                        }
                    ]
                }
            }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Error("failed to parse the test JSON: ", err)
	}

	if len(so.Properties["subobject1"].AdditionalProperties) != 1 {
		t.Error("expected 1 schemas in subobject3")
	}

	if len(so.Properties["subobject2"].AdditionalProperties) != 1 {
		t.Error("expected 1 schemas in subobject3")
	}

	if len(so.Properties["subobject3"].AdditionalProperties) != 3 {
		t.Error("expected 3 schemas in subobject3")
	}

	// Check that the types can be extracted into a map.
	types := so.ExtractTypes()

	if len(types) != 2 {
		t.Errorf("expected 2 types, the example and subobject, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the names of the types map to expected references.
	if _, ok := types["#/properties/subobject2"]; !ok {
		t.Errorf("was expecting to find the subobject2 type in the map under key #/properties/subobject3, available types were %s",
			strings.Join(getKeyNames(types), ", "))
	}
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

	if len(types) != 2 {
		t.Errorf("Expected 2 types, ProductSet and Product, but got %d types - %s", len(types),
			strings.Join(getKeyNames(types), ", "))
	}

	// Check that the keys of the types map to expected references.
	pss, ok := types["#"]
	if !ok {
		t.Errorf("Expected to find the '#' schema path, but available paths were %s",
			strings.Join(getKeyNames(types), ", "))
	}
	ps, ok := types["#/arrayitems"]
	if !ok {
		t.Errorf("Expected to find the '#/arrayitems' schema path, but available paths were %s",
			strings.Join(getKeyNames(types), ", "))
	}

	if pss.Title != "ProductSet" {
		t.Errorf("Expected the root schema's title to be 'ProductSet', but it was %s", pss.Title)
	}

	if ps.Title != "Product" {
		t.Errorf("Expected the array item's title to be 'Product', but it was %s", ps.Title)
	}

	if len(ps.Properties) != 4 {
		t.Errorf("Expected the Product schema to have 4 properties, but it had %d", len(ps.Properties))
	}

	tagType, _ := ps.Properties["tags"].Type()
	if tagType != "array" {
		t.Errorf("Expected the Tags property type to be 'array', but it was %s", tagType)
	}
}

func TestThatReferencesCanBeListed(t *testing.T) {
	s := `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "title": "Product set",
    "type": "array",
    "definitions": {
        "address": {
            "properties": {
                "houseName": { "type": "string" },
                "postcode": { "type": "string" }
            }
        }
    },
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
		t.Fatal("Failed to parse the test JSON: ", err)
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
*/

func TestThatPropertiesCanHaveMultipleTypes(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/schema#",
        "title": "root",
        "properties": {
            "name": {
                "type": [ "integer", "string" ]
            }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Fatal("It was not possible to unmarshal the schema:", err)
	}

	nameType, nameMultiple := so.Properties["name"].Type()
	if nameType != "integer" {
		t.Errorf("expected first value of property 'name' type to be 'integer', but was '%v'", nameType)
	}

	if !nameMultiple {
		t.Errorf("expected multiple types, but only returned one")
	}
}

func TestThatParsingInvalidValuesReturnsAnError(t *testing.T) {
	s := `{ " }`
	_, err := Parse(s)

	if err == nil {
		t.Fatal("Expected a parsing error, but got nil")
	}
}

func TestThatDefaultsCanBeParsed(t *testing.T) {
	s := `{
        "$schema": "http://json-schema.org/schema#",
        "title": "root",
        "properties": {
            "name": {
                "type": [ "integer", "string" ],
                "default":"Enrique"
            }
        }
    }`
	so, err := Parse(s)

	if err != nil {
		t.Fatal("It was not possible to unmarshal the schema:", err)
	}

	defaultValue := so.Properties["name"].Default
	if defaultValue != "Enrique" {
		t.Errorf("expected default value of property 'name' type to be 'Enrique', but was '%v'", defaultValue)
	}
}

func TestReturnedSchemaId(t *testing.T) {
	tests := []struct {
		input    *Schema
		expected string
	}{
		{
			input:    &Schema{},
			expected: "",
		},
		{
			input:    &Schema{ID06: "http://example.com/foo.json", ID04: "#foo"},
			expected: "http://example.com/foo.json",
		},
		{
			input:    &Schema{ID04: "#foo"},
			expected: "#foo",
		},
	}

	for idx, test := range tests {
		actual := test.input.ID()
		if actual != test.expected {
			t.Errorf("Test %d failed: For input \"%+v\", expected \"%s\", got \"%s\"", idx, test.input, test.expected, actual)
		}
	}
}
