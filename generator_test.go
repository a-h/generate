package generate

import (
	"strings"
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
		schema   jsonschema.Schema
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
		{
			input:    "#",
			expected: "TheRootName",
			schema: jsonschema.Schema{
				Title: "TheRootName",
			},
		},
	}

	for idx, test := range tests {
		actual := getStructName(test.input, &test.schema, 1)
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

	requiredFields := []string{"property2"}
	result, err := getFields("#", properties, lookupTypes, requiredFields)

	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, but got %d results", len(result))
	}

	testField(result["Property1"], "property1", "Property1", "string", false, t)
	testField(result["Property2"], "property2", "Property2", "*Address", true, t)
}

func TestFieldGenerationWithArrayReferences(t *testing.T) {
	properties := map[string]*jsonschema.Schema{
		"property1": &jsonschema.Schema{Type: "string"},
		"property2": &jsonschema.Schema{
			Type: "array",
			Items: &jsonschema.Schema{
				Reference: "#/definitions/address",
			},
		},
	}

	lookupTypes := map[string]*jsonschema.Schema{
		"#/definitions/address": &jsonschema.Schema{},
	}

	requiredFields := []string{"property2"}
	result, err := getFields("#", properties, lookupTypes, requiredFields)

	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, but got %d results", len(result))
	}

	testField(result["Property1"], "property1", "Property1", "string", false, t)
	testField(result["Property2"], "property2", "Property2", "[]Address", true, t)
}

func testField(actual Field, expectedJSONName string, expectedName string, expectedType string, expectedToBeRequired bool, t *testing.T) {
	if actual.JSONName != expectedJSONName {
		t.Errorf("JSONName - expected %s, got %s", expectedJSONName, actual.JSONName)
	}
	if actual.Name != expectedName {
		t.Errorf("Name - expected %s, got %s", expectedName, actual.Name)
	}
	if actual.Type != expectedType {
		t.Errorf("Type - expected %s, got %s", expectedType, actual.Type)
	}
	if actual.Required != expectedToBeRequired {
		t.Errorf("Required - expected %v, got %v", expectedToBeRequired, actual.Required)
	}
}

func TestNestedStructGeneration(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Example"
	root.Properties = map[string]*jsonschema.Schema{
		"property1": &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"subproperty1": &jsonschema.Schema{Type: "string"},
			},
		},
	}

	g := New(root)
	results, err := g.CreateStructs()

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Errorf("2 results should have been created, a root type and a type for the object 'property1' but %d structs were made", len(results))
	}

	if _, contains := results["Example"]; !contains {
		t.Errorf("The Example type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if _, contains := results["Property1"]; !contains {
		t.Errorf("The Property1 type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if results["Example"].Fields["Property1"].Type != "*Property1" {
		t.Errorf("Expected that the nested type property1 is generated as a struct, so the property type should be *Property1, but was %s.", results["Example"].Fields["Property1"].Type)
	}
}

func TestStructNameExtractor(t *testing.T) {
	m := make(map[string]Struct)
	m["name1"] = Struct{}
	m["name2"] = Struct{}

	names := getStructNamesFromMap(m)
	if len(names) != 2 {
		t.Error("Didn't extract all names from the map.")
	}

	if !contains(names, "name1") {
		t.Error("name1 was not extracted")
	}

	if !contains(names, "name2") {
		t.Error("name2 was not extracted")
	}
}

func getStructNamesFromMap(m map[string]Struct) []string {
	sn := make([]string, len(m))
	i := 0
	for k := range m {
		sn[i] = k
		i++
	}
	return sn
}

func TestStructGeneration(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "RootElement"
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
	results, err := g.CreateStructs()

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Error("2 results should have been created, a root type and an address")
	}
}

func TestArrayGeneration(t *testing.T) {
	root := &jsonschema.Schema{
		Title: "Array of Artists Example",
		Type:  "array",
		Items: &jsonschema.Schema{
			Title: "Artist",
			Type:  "object",
			Properties: map[string]*jsonschema.Schema{
				"name":      &jsonschema.Schema{Type: "string"},
				"birthyear": &jsonschema.Schema{Type: "number"},
			},
		},
	}

	g := New(root)
	results, err := g.CreateStructs()

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected one struct should have been generated, but %d have been generated.", len(results))
	}

	artistStruct, ok := results["Artist"]

	if !ok {
		t.Errorf("Expected Name to be Artist, that wasn't found, but the struct contains \"%+v\"", results)
	}

	if len(artistStruct.Fields) != 2 {
		t.Errorf("Expected the fields to be birtyear and name, but %d fields were found.", len(artistStruct.Fields))
	}

	if _, ok := artistStruct.Fields["Name"]; !ok {
		t.Errorf("Expected to find a Name field, but one was not found.")
	}

	if _, ok := artistStruct.Fields["Birthyear"]; !ok {
		t.Errorf("Expected to find a Birthyear field, but one was not found.")
	}
}

func TestNestedArrayGeneration(t *testing.T) {
	root := &jsonschema.Schema{
		Title: "Favourite Bars",
		Type:  "object",
		Properties: map[string]*jsonschema.Schema{
			"barName": &jsonschema.Schema{Type: "string"},
			"cities": &jsonschema.Schema{
				Type: "array",
				Items: &jsonschema.Schema{
					Title: "City",
					Properties: map[string]*jsonschema.Schema{
						"name":    &jsonschema.Schema{Type: "string"},
						"country": &jsonschema.Schema{Type: "string"},
					},
				},
			},
			"tags": &jsonschema.Schema{
				Type:  "array",
				Items: &jsonschema.Schema{Type: "string"},
			},
		},
	}

	g := New(root)
	results, err := g.CreateStructs()

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected two structs to be generated - 'Favourite Bars' and 'City', but %d have been generated.", len(results))
	}

	fbStruct, ok := results["FavouriteBars"]

	if !ok {
		t.Errorf("FavouriteBars struct was not found. The results were %+v", results)
	}

	if _, ok := fbStruct.Fields["BarName"]; !ok {
		t.Errorf("Expected to find the BarName field, but didn't. The struct is %+v", fbStruct)
	}

	if f, ok := fbStruct.Fields["Cities"]; !ok {
		t.Errorf("Expected to find the Cities field on the FavouriteBars, but didn't. The struct is %+v", fbStruct)

		if f.Type != "City" {
			t.Errorf("Expected to find that the Cities array was of type City, but it was of %s", f.Type)
		}
	}

	if f, ok := fbStruct.Fields["Tags"]; !ok {
		t.Errorf("Expected to find the Tags field on the FavouriteBars, but didn't. The struct is %+v", fbStruct)

		if f.Type != "array" {
			t.Errorf("Expected to find that the Tags array was of type array, but it was of %s", f.Type)
		}
	}

	cityStruct, ok := results["City"]

	if !ok {
		t.Error("City struct was not found.")
	}

	if _, ok := cityStruct.Fields["Name"]; !ok {
		t.Errorf("Expected to find the Name field on the City struct, but didn't. The struct is %+v", cityStruct)
	}

	if _, ok := cityStruct.Fields["Country"]; !ok {
		t.Errorf("Expected to find the Country field on the City struct, but didn't. The struct is %+v", cityStruct)
	}
}

func TestThatJavascriptKeyNamesCanBeConvertedToValidGoNames(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    string
	}{
		{
			description: "Camel case is converted to pascal case.",
			input:       "camelCase",
			expected:    "CamelCase",
		},
		{
			description: "Spaces are stripped.",
			input:       "Contains space",
			expected:    "ContainsSpace",
		},
		{
			description: "Hyphens are stripped.",
			input:       "key-name",
			expected:    "KeyName",
		},
		{
			description: "Underscores are stripped.",
			input:       "key_name",
			expected:    "KeyName",
		},
		{
			description: "Periods are stripped.",
			input:       "a.b.c",
			expected:    "ABC",
		},
	}

	for _, test := range tests {
		actual := getGolangName(test.input)

		if test.expected != actual {
			t.Errorf("For test '%s', for input '%s' expected '%s' but got '%s'.", test.description, test.input, test.expected, actual)
		}
	}
}
