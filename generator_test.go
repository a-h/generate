package generate

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
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

func TestFieldGeneration(t *testing.T) {
	properties := map[string]*Schema{
		"property1": {TypeValue: "string"},
		"property2": {Reference: "#/definitions/address"},
		// test sub-objects with properties or additionalProperties
		"property3": {TypeValue: "object", Title: "SubObj1", Properties: map[string]*Schema{"name": {TypeValue: "string"}}},
		"property4": {TypeValue: "object", Title: "SubObj2", AdditionalProperties: &AdditionalProperties{TypeValue: "integer"}},
		// test sub-objects with properties composed of objects
		"property5": {TypeValue: "object", Title: "SubObj3", Properties: map[string]*Schema{"SubObj3a": {TypeValue: "object", Properties: map[string]*Schema{"subproperty1": {TypeValue: "integer"}}}}},
		// test sub-objects with additionalProperties composed of objects
		"property6": {TypeValue: "object", Title: "SubObj4", AdditionalProperties: &AdditionalProperties{TypeValue: "object", Title: "SubObj4a", Properties: map[string]*Schema{"subproperty1": {TypeValue: "integer"}}}},
		// test sub-objects without titles
		"property7": {TypeValue: "object"},
		// test sub-objects with properties AND additionalProperties
		"property8": {TypeValue: "object", Title: "SubObj5", Properties: map[string]*Schema{"name": {TypeValue: "string"}}, AdditionalProperties: &AdditionalProperties{TypeValue: "integer"}},
	}

	requiredFields := []string{"property2"}

	root := Schema{
		SchemaType: "http://localhost",
		Title:      "TestFieldGeneration",
		TypeValue:  "object",
		Properties: properties,
		Definitions: map[string]*Schema{
			"address": {TypeValue: "object"},
		},
		Required: requiredFields,
	}
	root.Init()
	g := New(&root)
	err := g.CreateTypes()

	// Output(os.Stderr, g, "test")

	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(g.Structs) != 8 {
		t.Errorf("Expected 8 results, but got %d results", len(g.Structs))
	}

	testField(g.Structs["TestFieldGeneration"].Fields["Property1"], "property1", "Property1", "string", false, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property2"], "property2", "Property2", "*Address", true, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property3"], "property3", "Property3", "*SubObj1", false, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property4"], "property4", "Property4", "map[string]int", false, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property5"], "property5", "Property5", "*SubObj3", false, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property6"], "property6", "Property6", "map[string]*SubObj4a", false, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property7"], "property7", "Property7", "*Property7", false, t)
	testField(g.Structs["TestFieldGeneration"].Fields["Property8"], "property8", "Property8", "*SubObj5", false, t)

	testField(g.Structs["SubObj1"].Fields["Name"], "name", "Name", "string", false, t)
	testField(g.Structs["SubObj3"].Fields["SubObj3a"], "SubObj3a", "SubObj3a", "*SubObj3a", false, t)
	testField(g.Structs["SubObj4a"].Fields["Subproperty1"], "subproperty1", "Subproperty1", "int", false, t)

	testField(g.Structs["SubObj5"].Fields["Name"], "name", "Name", "string", false, t)
	testField(g.Structs["SubObj5"].Fields["AdditionalProperties"], "-", "AdditionalProperties", "map[string]int", false, t)

	if strct, ok := g.Structs["Property7"]; !ok {
		t.Fatal("Property7 wasn't generated")
	} else {
		if len(strct.Fields) != 0 {
			t.Fatal("Property7 expected 0 fields")
		}
	}
}

func TestFieldGenerationWithArrayReferences(t *testing.T) {
	properties := map[string]*Schema{
		"property1": {TypeValue: "string"},
		"property2": {
			TypeValue: "array",
			Items: &Schema{
				Reference: "#/definitions/address",
			},
		},
		"property3": {
			TypeValue: "array",
			Items: &Schema{
				TypeValue:            "object",
				AdditionalProperties: (*AdditionalProperties)(&Schema{TypeValue: "integer"}),
			},
		},
		"property4": {
			TypeValue: "array",
			Items: &Schema{
				Reference: "#/definitions/outer",
			},
		},
	}

	requiredFields := []string{"property2"}

	root := Schema{
		SchemaType: "http://localhost",
		Title:      "TestFieldGenerationWithArrayReferences",
		TypeValue:  "object",
		Properties: properties,
		Definitions: map[string]*Schema{
			"address": {TypeValue: "object"},
			"outer":   {TypeValue: "array", Items: &Schema{Reference: "#/definitions/inner"}},
			"inner":   {TypeValue: "object"},
		},
		Required: requiredFields,
	}
	root.Init()

	g := New(&root)
	err := g.CreateTypes()

	//Output(os.Stderr, g, "test")

	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(g.Structs) != 3 {
		t.Errorf("Expected 3 results, but got %d results", len(g.Structs))
	}

	testField(g.Structs["TestFieldGenerationWithArrayReferences"].Fields["Property1"], "property1", "Property1", "string", false, t)
	testField(g.Structs["TestFieldGenerationWithArrayReferences"].Fields["Property2"], "property2", "Property2", "[]*Address", true, t)
	testField(g.Structs["TestFieldGenerationWithArrayReferences"].Fields["Property3"], "property3", "Property3", "[]map[string]int", false, t)
	testField(g.Structs["TestFieldGenerationWithArrayReferences"].Fields["Property4"], "property4", "Property4", "[][]*Inner", false, t)
}

func testField(actual Field, expectedJSONName string, expectedName string, expectedType string, expectedToBeRequired bool, t *testing.T) {
	if actual.JSONName != expectedJSONName {
		t.Errorf("JSONName - expected \"%s\", got \"%s\"", expectedJSONName, actual.JSONName)
	}
	if actual.Name != expectedName {
		t.Errorf("Name - expected \"%s\", got \"%s\"", expectedName, actual.Name)
	}
	if actual.Type != expectedType {
		t.Errorf("Type - expected \"%s\", got \"%s\"", expectedType, actual.Type)
	}
	if actual.Required != expectedToBeRequired {
		t.Errorf("Required - expected \"%v\", got \"%v\"", expectedToBeRequired, actual.Required)
	}
}

func TestNestedStructGeneration(t *testing.T) {
	root := &Schema{}
	root.Title = "Example"
	root.Properties = map[string]*Schema{
		"property1": {
			TypeValue: "object",
			Properties: map[string]*Schema{
				"subproperty1": {TypeValue: "string"},
			},
		},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	//Output(os.Stderr, g, "test")

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
	root := &Schema{}
	root.Title = "Example"
	root.Properties = map[string]*Schema{
		"property1": {
			TypeValue: "object",
			Properties: map[string]*Schema{
				"nestedproperty1": {TypeValue: "string"},
			},
		},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

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
	root := &Schema{}
	root.Title = "RootElement"
	root.Definitions = make(map[string]*Schema)
	root.Definitions["address"] = &Schema{
		Properties: map[string]*Schema{
			"address1": {TypeValue: "string"},
			"zip":      {TypeValue: "number"},
		},
	}
	root.Properties = map[string]*Schema{
		"property1": {TypeValue: "string"},
		"property2": {Reference: "#/definitions/address"},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Error("2 results should have been created, a root type and an address")
	}
}

func TestArrayGeneration(t *testing.T) {
	root := &Schema{
		Title:     "Array of Artists Example",
		TypeValue: "array",
		Items: &Schema{
			Title:     "Artist",
			TypeValue: "object",
			Properties: map[string]*Schema{
				"name":      {TypeValue: "string"},
				"birthyear": {TypeValue: "number"},
			},
		},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

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
	root := &Schema{
		Title:     "Favourite Bars",
		TypeValue: "object",
		Properties: map[string]*Schema{
			"barName": {TypeValue: "string"},
			"cities": {
				TypeValue: "array",
				Items: &Schema{
					Title:     "City",
					TypeValue: "object",
					Properties: map[string]*Schema{
						"name":    {TypeValue: "string"},
						"country": {TypeValue: "string"},
					},
				},
			},
			"tags": {
				TypeValue: "array",
				Items:     &Schema{TypeValue: "string"},
			},
		},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

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
	root1 := &Schema{
		Title: "Root1Element",
		ID06:  "http://example.com/schema/root1",
		Properties: map[string]*Schema{
			"property1": {Reference: "root2#/definitions/address"},
		},
	}

	root2 := &Schema{
		Title: "Root2Element",
		ID06:  "http://example.com/schema/root2",
		Properties: map[string]*Schema{
			"property1": {Reference: "#/definitions/address"},
		},
		Definitions: map[string]*Schema{
			"address": {
				Properties: map[string]*Schema{
					"address1": {TypeValue: "string"},
					"zip":      {TypeValue: "number"},
				},
			},
		},
	}

	root1.Init()
	root2.Init()

	g := New(root1, root2)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

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
	root := &Schema{}
	root.Title = "Array without defined item"
	root.Properties = map[string]*Schema{
		"name": {TypeValue: "string"},
		"repositories": {
			TypeValue: "array",
		},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

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
	root := &Schema{}
	root.Title = "Multiple possible types"
	root.Properties = map[string]*Schema{
		"name": {TypeValue: []interface{}{"string", "integer"}},
	}

	root.Init()

	g := New(root)
	err := g.CreateTypes()
	results := g.Structs

	// Output(os.Stderr, g, "test")

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

func TestTypeAliases(t *testing.T) {
	tests := []struct {
		gotype           string
		input            *Schema
		structs, aliases int
	}{
		{
			gotype:  "string",
			input:   &Schema{TypeValue: "string"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype:  "int",
			input:   &Schema{TypeValue: "integer"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype:  "bool",
			input:   &Schema{TypeValue: "boolean"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype: "[]*Foo",
			input: &Schema{TypeValue: "array",
				Items: &Schema{
					TypeValue: "object",
					Title:     "foo",
					Properties: map[string]*Schema{
						"nestedproperty": {TypeValue: "string"},
					},
				}},
			structs: 1,
			aliases: 1,
		},
		{
			gotype:  "[]interface{}",
			input:   &Schema{TypeValue: "array"},
			structs: 0,
			aliases: 1,
		},
		{
			gotype: "map[string]string",
			input: &Schema{
				TypeValue:            "object",
				AdditionalProperties: (*AdditionalProperties)(&Schema{TypeValue: "string"}),
			},
			structs: 0,
			aliases: 1,
		},
		{
			gotype: "map[string]interface{}",
			input: &Schema{
				TypeValue:            "object",
				AdditionalProperties: (*AdditionalProperties)(&Schema{TypeValue: []interface{}{"string", "integer"}}),
			},
			structs: 0,
			aliases: 1,
		},
	}

	for _, test := range tests {
		test.input.Init()

		g := New(test.input)
		err := g.CreateTypes()
		structs := g.Structs
		aliases := g.Aliases

		// Output(os.Stderr, g, "test")

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
