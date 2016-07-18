package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	"github.com/a-h/generate"
	"github.com/a-h/generate/jsonschema"
)

var i = flag.String("i", "", "The input JSON Schema file.")

func main() {
	flag.Parse()

	b, err := ioutil.ReadFile(*i)

	if err != nil {
		log.Fatalf("Failed to read the input file with error %s", err.Error())
	}

	schema, err := jsonschema.Parse(string(b))

	if err != nil {
		log.Fatalf("Failed to parse the input JSON schema with error %s", err.Error())
	}

	g := generate.New(schema)

	structs := g.CreateStructs()

	//TODO: Use templates.
	fmt.Println("package main")

	for _, k := range getOrderedStructNames(structs) {
		s := structs[k]

		fmt.Println("")
		fmt.Printf("type %s struct {\n", s.Name)

		for _, fieldKey := range getOrderedFieldNames(s.Fields) {
			f := s.Fields[fieldKey]
			fmt.Printf("  %s %s `json:\"%s\"`\n", f.Name, f.Type, f.JSONName)
		}

		fmt.Println("}")
	}
}

func getOrderedFieldNames(m map[string]generate.Field) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	sort.Strings(keys)
	return keys
}

func getOrderedStructNames(m map[string]generate.Struct) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	sort.Strings(keys)
	return keys
}
