package generate

import (
	"encoding/json"
	"net/url"
	"reflect"
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
			input:    "/definitions/address",
			expected: "Address",
		},
		{
			input:    "/Example",
			expected: "Example",
		},
		{
			input:    "/Example",
			expected: "Example",
			schema: jsonschema.Schema{
				NameCount: 1,
			},
		},
		{
			input:    "/Example",
			expected: "Example2",
			schema: jsonschema.Schema{
				NameCount: 2,
			},
		},
		{
			input:    "",
			expected: "TheRootName",
			schema: jsonschema.Schema{
				Title: "TheRootName",
			},
		},
	}

	for idx, test := range tests {
		actual := getTypeName(&url.URL{Fragment: test.input}, &test.schema, 1)
		if actual != test.expected {
			t.Errorf("Test %d failed: For input \"%s\", expected \"%s\", got \"%s\"", idx, test.input, test.expected, actual)
		}
	}
}

func TestFieldGeneration(t *testing.T) {
	properties := map[string]*jsonschema.Schema{
		"property1": {TypeValue: "string"},
		"property2": {Reference: "#/definitions/address"},
		"property3": {TypeValue: "object", AdditionalProperties: []*jsonschema.Schema{{TypeValue: "integer"}}},
		"property4": {TypeValue: "object", AdditionalProperties: []*jsonschema.Schema{{TypeValue: "integer"}, {TypeValue: "integer"}}},
		"property5": {TypeValue: "object", AdditionalProperties: []*jsonschema.Schema{{TypeValue: "object", Properties: map[string]*jsonschema.Schema{"subproperty1": {TypeValue: "integer"}}}}},
		"property6": {TypeValue: "object", AdditionalProperties: []*jsonschema.Schema{{TypeValue: "object", Properties: map[string]*jsonschema.Schema{"subproperty1": {TypeValue: "integer"}}}}},
	}

	lookupTypes := map[string]*jsonschema.Schema{
		"#/definitions/address":  {},
		"#/properties/property5": properties["property5"].AdditionalProperties[0],
	}

	requiredFields := []string{"property2"}
	result, err := getFields(&url.URL{}, properties, lookupTypes, requiredFields)

	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(result) != 6 {
		t.Errorf("Expected 6 results, but got %d results", len(result))
	}

	testField(result["Property1"], "property1", "Property1", "string", false, t)
	testField(result["Property2"], "property2", "Property2", "*Address", true, t)
	testField(result["Property3"], "property3", "Property3", "map[string]int", false, t)
	testField(result["Property4"], "property4", "Property4", "map[string]interface{}", false, t)
	testField(result["Property5"], "property5", "Property5", "map[string]*Property5", false, t)
	testField(result["Property6"], "property6", "Property6", "map[string]*undefined", false, t)
}

func TestFieldGenerationWithArrayReferences(t *testing.T) {
	properties := map[string]*jsonschema.Schema{
		"property1": {TypeValue: "string"},
		"property2": {
			TypeValue: "array",
			Items: &jsonschema.Schema{
				Reference: "#/definitions/address",
			},
		},
		"property3": {
			TypeValue: "array",
			Items: &jsonschema.Schema{
				TypeValue:            "object",
				AdditionalProperties: []*jsonschema.Schema{{TypeValue: "integer"}},
			},
		},
	}

	lookupTypes := map[string]*jsonschema.Schema{
		"#/definitions/address": {},
	}

	requiredFields := []string{"property2"}
	result, err := getFields(&url.URL{}, properties, lookupTypes, requiredFields)

	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 results, but got %d results", len(result))
	}

	testField(result["Property1"], "property1", "Property1", "string", false, t)
	testField(result["Property2"], "property2", "Property2", "[]*Address", true, t)
	testField(result["Property3"], "property3", "Property3", "[]map[string]int", false, t)
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
		"property1": {
			TypeValue: "object",
			Properties: map[string]*jsonschema.Schema{
				"subproperty1": {TypeValue: "string"},
			},
		},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

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

func TestEmptyNestedStructGeneration(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Example"
	root.Properties = map[string]*jsonschema.Schema{
		"property1": {
			TypeValue: "object",
			Properties: map[string]*jsonschema.Schema{
				"nestedproperty1": {TypeValue: "string"},
			},
		},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

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
			"address1": {TypeValue: "string"},
			"zip":      {TypeValue: "number"},
		},
	}
	root.Properties = map[string]*jsonschema.Schema{
		"property1": {TypeValue: "string"},
		"property2": {Reference: "#/definitions/address"},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Error("2 results should have been created, a root type and an address")
	}
}

func TestArrayGeneration(t *testing.T) {
	root := &jsonschema.Schema{
		Title:     "Array of Artists Example",
		TypeValue: "array",
		Items: &jsonschema.Schema{
			Title:     "Artist",
			TypeValue: "object",
			Properties: map[string]*jsonschema.Schema{
				"name":      {TypeValue: "string"},
				"birthyear": {TypeValue: "number"},
			},
		},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

	if err != nil {
		t.Fatal("Failed to create structs: ", err)
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
		Title:     "Favourite Bars",
		TypeValue: "object",
		Properties: map[string]*jsonschema.Schema{
			"barName": {TypeValue: "string"},
			"cities": {
				TypeValue: "array",
				Items: &jsonschema.Schema{
					Title:     "City",
					TypeValue: "object",
					Properties: map[string]*jsonschema.Schema{
						"name":    {TypeValue: "string"},
						"country": {TypeValue: "string"},
					},
				},
			},
			"tags": {
				TypeValue: "array",
				Items:     &jsonschema.Schema{TypeValue: "string"},
			},
		},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

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

	f, ok := fbStruct.Fields["Cities"]
	if !ok {
		t.Errorf("Expected to find the Cities field on the FavouriteBars, but didn't. The struct is %+v", fbStruct)
	}
	if f.Type != "[]*City" {
		t.Errorf("Expected to find that the Cities array was of type *City, but it was of %s", f.Type)
	}

	f, ok = fbStruct.Fields["Tags"]
	if !ok {
		t.Errorf("Expected to find the Tags field on the FavouriteBars, but didn't. The struct is %+v", fbStruct)
	}

	if f.Type != "[]string" {
		t.Errorf("Expected to find that the Tags array was of type string, but it was of %s", f.Type)
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

func TestMultipleSchemaStructGeneration(t *testing.T) {
	root1 := &jsonschema.Schema{
		Title: "Root1Element",
		ID06:  "http://example.com/schema/root1",
		Properties: map[string]*jsonschema.Schema{
			"property1": {Reference: "root2#/definitions/address"},
		},
	}

	root2 := &jsonschema.Schema{
		Title: "Root2Element",
		ID06:  "http://example.com/schema/root2",
		Properties: map[string]*jsonschema.Schema{
			"property1": {Reference: "#/definitions/address"},
		},
		Definitions: map[string]*jsonschema.Schema{
			"address": {
				Properties: map[string]*jsonschema.Schema{
					"address1": {TypeValue: "string"},
					"zip":      {TypeValue: "number"},
				},
			},
		},
	}

	g := New(root1, root2)
	results, _, err := g.CreateTypes()

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 3 {
		t.Errorf("3 results should have been created, 2 root types and an address, but got %v", getStructNamesFromMap(results))
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
		{
			description: "Colons are stripped.",
			input:       "a:b",
			expected:    "AB",
		},
		{
			description: "GT and LT are stripped.",
			input:       "a<b>",
			expected:    "AB",
		},
		{
			description: "Not allowed to start with a number.",
			input:       "123ABC",
			expected:    "_123ABC",
		},
	}

	for _, test := range tests {
		actual := getGolangName(test.input)

		if test.expected != actual {
			t.Errorf("For test '%s', for input '%s' expected '%s' but got '%s'.", test.description, test.input, test.expected, actual)
		}
	}
}

func TestThatArraysWithoutDefinedItemTypesAreGeneratedAsEmptyInterfaces(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Array without defined item"
	root.Properties = map[string]*jsonschema.Schema{
		"name": {TypeValue: "string"},
		"repositories": {
			TypeValue: "array",
		},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

	if err != nil {
		t.Errorf("Error generating structs: %v", err)
	}

	if _, contains := results["ArrayWithoutDefinedItem"]; !contains {
		t.Errorf("The ArrayWithoutDefinedItem type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if o, ok := results["ArrayWithoutDefinedItem"]; ok {
		if f, ok := o.Fields["Repositories"]; ok {
			if f.Type != "[]interface{}" {
				t.Errorf("Since the schema doesn't include a type for the array items, the property type should be []interface{}, but was %s.", f.Type)
			}
		} else {
			t.Errorf("Expected the ArrayWithoutDefinedItem type to have a Repostitories field, but none was found.")
		}
	}
}

func TestThatTypesWithMultipleDefinitionsAreGeneratedAsEmptyInterfaces(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Multiple possible types"
	root.Properties = map[string]*jsonschema.Schema{
		"name": {TypeValue: []interface{}{"string", "integer"}},
	}

	g := New(root)
	results, _, err := g.CreateTypes()

	if err != nil {
		t.Errorf("Error generating structs: %v", err)
	}

	if _, contains := results["MultiplePossibleTypes"]; !contains {
		t.Errorf("The MultiplePossibleTypes type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if o, ok := results["MultiplePossibleTypes"]; ok {
		if f, ok := o.Fields["Name"]; ok {
			if f.Type != "interface{}" {
				t.Errorf("Since the schema has multiple types for the item, the property type should be []interface{}, but was %s.", f.Type)
			}
		} else {
			t.Errorf("Expected the MultiplePossibleTypes type to have a Name field, but none was found.")
		}
	}
}

func TestThatUnmarshallingIsPossible(t *testing.T) {
	// {
	//     "$schema": "http://json-schema.org/draft-04/schema#",
	//     "name": "Example",
	//     "type": "object",
	//     "properties": {
	//         "name": {
	//             "type": ["object", "array", "integer"],
	//             "description": "name"
	// 		}
	//     }
	// }

	tests := []struct {
		name     string
		input    string
		expected Root
	}{
		{
			name:  "map",
			input: `{ "name": { "key": "value" } }`,
			expected: Root{
				Name: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name:  "array",
			input: `{ "name": [ "a", "b" ] }`,
			expected: Root{
				Name: []interface{}{"a", "b"},
			},
		},
		{
			name:  "integer",
			input: `{ "name": 1 }`,
			expected: Root{
				Name: 1.0, // can't determine whether it's a float or integer without additional info
			},
		},
	}

	for _, test := range tests {
		var actual Root
		err := json.Unmarshal([]byte(test.input), &actual)
		if err != nil {
			t.Errorf("%s: error unmarshalling: %v", test.name, err)
		}

		expectedType := reflect.TypeOf(test.expected.Name)
		actualType := reflect.TypeOf(actual.Name)
		if expectedType != actualType {
			t.Errorf("expected Name to be of type %v, but got %v", expectedType, actualType)
		}
	}
}

func TestThatRootTypeKeyIsCorrectlyAssessed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "URL without fragment",
			input:    "http://example.com/schema",
			expected: true,
		},
		{
			name:     "URL with fragment",
			input:    "http://example.com/schema#/definitions/foo",
			expected: false,
		},
		{
			name:     "simple ID without fragment",
			input:    "/Test",
			expected: true,
		},
		{
			name:     "simple ID with fragment",
			input:    "/Test#/definitions/foo",
			expected: false,
		},
		{
			name:     "no ID",
			input:    "#",
			expected: true,
		},
		{
			name:     "empty",
			input:    "",
			expected: true,
		},
	}

	for _, test := range tests {
		key, err := url.Parse(test.input)
		if err != nil {
			t.Fatal(err)
		}

		actual := isRootSchemaKey(key)
		if actual != test.expected {
			t.Errorf("Test %q failed: for input %q, expected %t, got %t", test.name, test.input, test.expected, actual)
		}
	}
}

func TestTypeAliases(t *testing.T) {
	tests := []struct {
		gotype           string
		input            *jsonschema.Schema
		structs, aliases int
	}{
		{
			gotype:  "string",
			input:   &jsonschema.Schema{TypeValue: "string"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype:  "int",
			input:   &jsonschema.Schema{TypeValue: "integer"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype:  "bool",
			input:   &jsonschema.Schema{TypeValue: "boolean"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype: "[]*Foo",
			input: &jsonschema.Schema{TypeValue: "array",
				Items: &jsonschema.Schema{
					TypeValue: "object",
					Title:     "foo",
					Properties: map[string]*jsonschema.Schema{
						"nestedproperty": {TypeValue: "string"},
					},
				}},
			structs: 1,
			aliases: 1,
		},
		{
			gotype:  "[]interface{}",
			input:   &jsonschema.Schema{TypeValue: "array"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype: "map[string]string",
			input: &jsonschema.Schema{
				TypeValue:            "object",
				AdditionalProperties: []*jsonschema.Schema{{TypeValue: "string"}},
			},
			structs: 0,
			aliases: 1,
		},
		{
			gotype: "map[string]interface{}",
			input: &jsonschema.Schema{
				TypeValue:            "object",
				AdditionalProperties: []*jsonschema.Schema{{TypeValue: []interface{}{"string", "integer"}}},
			},
			structs: 0,
			aliases: 1,
		},
	}

	for _, test := range tests {
		g := New(test.input)
		structs, aliases, err := g.CreateTypes()
		if err != nil {
			t.Fatal(err)
		}

		if len(structs) != test.structs {
			t.Errorf("Expected %d structs, got %d", test.structs, len(structs))
		}

		if len(aliases) != test.aliases {
			t.Errorf("Expected %d type aliases, got %d", test.aliases, len(aliases))
		}

		if test.gotype != aliases["Root"].Type {
			t.Errorf("Expected Root type %q, got %q", test.gotype, aliases["Root"].Type)
		}
	}
}

// Root is an example of a generated type.
type Root struct {
	Name interface{} `json:"name,omitempty"`
}
