package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/a-h/generate"
	"github.com/a-h/generate/jsonschema"
)

var (
	i = flag.String("i", "", "The input JSON Schema file.")
	o = flag.String("o", "", "The output file for the schema.")
	p = flag.String("p", "main", "The package that the structs are created in.")
)

func main() {
	flag.Parse()

	b, err := ioutil.ReadFile(*i)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read the input file with error ", err)
		return
	}

	schema, err := jsonschema.Parse(string(b))

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse the input JSON schema with error ", err)
		return
	}

	g := generate.New(schema)

	structs, err := g.CreateStructs()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failure generating structs: ", err)
	}

	var w io.Writer

	if *o == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(*o)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening output file: ", err)
			return
		}
	}

	output(w, structs)
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

func output(w io.Writer, structs map[string]generate.Struct) {
	//TODO: Use templates.
	fmt.Fprintf(w, "package %v\n", *p)

	for _, k := range getOrderedStructNames(structs) {
		s := structs[k]

		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "type %s struct {\n", s.Name)

		for _, fieldKey := range getOrderedFieldNames(s.Fields) {
			f := s.Fields[fieldKey]

			// Only apply omitempty if the field is not required.
			omitempty := ",omitempty"
			if f.Required {
				omitempty = ""
			}

			fmt.Fprintf(w, "  %s %s `json:\"%s%s\"`\n", f.Name, f.Type, f.JSONName, omitempty)
		}

		fmt.Fprintln(w, "}")
	}
}
