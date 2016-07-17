package generate

import (
	"log"
	"testing"

	"github.com/a-h/generate/jsonschema"
)

func TestThatCapitalisationOccursCorrectly(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "ssd",
			expected: "Ssd",
		},
		{
			input:    "f",
			expected: "F",
		},
		{
			input:    "fishPaste",
			expected: "FishPaste",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "F",
			expected: "F",
		},
	}

	for idx, test := range tests {
		actual := capitaliseFirstLetter(test.input)
		if actual != test.expected {
			t.Errorf("Test %d failed: For input \"%s\", expected \"%s\", got \"%s\"", idx, test.input, test.expected, actual)
		}
	}
}

func TestThatStructsAreNamedWell(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "#/definitions/address",
			expected: "Address",
		},
		{
			input:    "#/Example",
			expected: "Example",
		},
	}

	for idx, test := range tests {
		actual := getStructName(test.input)
		if actual != test.expected {
			t.Errorf("Test %d failed: For input \"%s\", expected \"%s\", got \"%s\"", idx, test.input, test.expected, actual)
		}
	}
}

func TestFieldGeneration(t *testing.T) {
	properties := map[string]*jsonschema.Schema{
		"property1": &jsonschema.Schema{Type: "string"},
		"property2": &jsonschema.Schema{Reference: "#/definitions/address"},
	}

	lookupTypes := map[string]*jsonschema.Schema{
		"#/definitions/address": &jsonschema.Schema{},
	}

	result := getFields(properties, lookupTypes)

	if len(result) != 2 {
		t.Errorf("Expected 2 results, but got %d results", len(result))
	}

	testField(result["Property1"], "property1", "Property1", "string", t)
	testField(result["Property2"], "property2", "Property2", "Address", t)
}

func testField(actual Field, expectedJSONName string, expectedName string, expectedType string, t *testing.T) {
	if actual.JSONName != expectedJSONName {
		t.Errorf("JSONName - expected %s, got %s", expectedJSONName, actual.JSONName)
	}
	if actual.Name != expectedName {
		t.Errorf("Name - expected %s, got %s", expectedName, actual.Name)
	}
	if actual.Type != expectedType {
		t.Errorf("Type - expected %s, got %s", expectedType, actual.Type)
	}
}

func TestStructGeneration(t *testing.T) {
	root := &jsonschema.Root{}
	root.Definitions = make(map[string]*jsonschema.Schema)
	root.Definitions["address"] = &jsonschema.Schema{
		Properties: map[string]*jsonschema.Schema{
			"address1": &jsonschema.Schema{Type: "string"},
			"zip":      &jsonschema.Schema{Type: "number"},
		},
	}
	root.Properties = map[string]*jsonschema.Schema{
		"property1": &jsonschema.Schema{Type: "string"},
		"property2": &jsonschema.Schema{Reference: "#/definitions/address"},
	}

	g := New(root)
	results := g.CreateStructs()

	if len(results) != 2 {
		t.Error("2 results should have been created, a root type and an address")
	}

	for _, v := range results {
		log.Print(v)
	}
}
