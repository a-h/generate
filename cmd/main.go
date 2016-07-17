package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

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

	for _, s := range structs {
		fmt.Println("")
		fmt.Printf("type %s struct {\n", s.Name)

		for _, f := range s.Fields {
			fmt.Printf("  %s %s `json:\"%s\"`\n", f.Name, f.Type, f.JSONName)
		}

		fmt.Println("}")
	}
}
