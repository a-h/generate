package generate

import (
	"sort"
	"io"
	"fmt"
	"strings"
)

func getOrderedFieldNames(m map[string]Field) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	sort.Strings(keys)
	return keys
}

func getOrderedStructNames(m map[string]Struct) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	sort.Strings(keys)
	return keys
}

// generate code and write to w
func Output(w io.Writer, g *Generator, pkg string) {
	structs := g.Structs
	aliases := g.Aliases

	fmt.Fprintln(w, "// Code  by schema- DO NOT EDIT.")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "package %v\n", pkg)

	willEmitCode := false
	for _, v := range structs {
		if v.AdditionalValueType != "" {
			willEmitCode = true
		}
	}

	if willEmitCode {
		fmt.Fprintf(w, `
import (
	"fmt"
	"encoding/json"
	"bytes"
)
`)
	}

	for _, k := range getOrderedFieldNames(aliases) {
		a := aliases[k]

		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "// %s\n", a.Name)
		fmt.Fprintf(w, "type %s %s\n", a.Name, a.Type)
	}

	for _, k := range getOrderedStructNames(structs) {
		s := structs[k]

		fmt.Fprintln(w, "")
		outputNameAndDescriptionComment(s.Name, s.Description, w)
		fmt.Fprintf(w, "type %s struct {\n", s.Name)

		for _, fieldKey := range getOrderedFieldNames(s.Fields) {
			f := s.Fields[fieldKey]

			// Only apply omitempty if the field is not required.
			omitempty := ",omitempty"
			if f.Required {
				omitempty = ""
			}

			if f.Comment != "" {
				fmt.Fprintf(w, "  // %s\n", f.Comment)
			}

			fmt.Fprintf(w, "  %s %s `json:\"%s%s\"`\n", f.Name, f.Type, f.JSONName, omitempty)
		}

		fmt.Fprintln(w, "}")

		if s.AdditionalValueType != "" {
			emitMarshalCode(w, s)
			emitUnmarshalCode(w, s)
		}
	}
}

func emitMarshalCode(w io.Writer, s Struct) {
	fmt.Fprintf(w,
		`
func (strct *%s) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	comma := false
`, s.Name)
	// Marshal all the defined fields
	for _, fieldKey := range getOrderedFieldNames(s.Fields) {
		f := s.Fields[fieldKey]
		if f.JSONName == "-" {
			continue
		}
		fmt.Fprintf(w,
			`    // Marshal the "%s" field
    if comma { 
        buf.WriteString(",") 
    }
    buf.WriteString("\"%s\": ")
	if tmp, err := json.Marshal(strct.%s); err != nil {
		return nil, err
 	} else {
 		buf.Write(tmp)
	}
	comma = true
`, f.JSONName, f.JSONName, f.Name)
	}
	fmt.Fprintf(w, "    // Marshal any additional Properties\n")
	// Marshal any additional Properties
	fmt.Fprintf(w, `    for k, v := range strct.AdditionalProperties {
		if comma {
			buf.WriteString(",")
		}
        buf.WriteString(fmt.Sprintf("\"%%s\":", k))
		if tmp, err := json.Marshal(v); err != nil {
			return nil, err
		} else {
			buf.Write(tmp)
		}
        comma = true
	}

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}
`)
}

func emitUnmarshalCode(w io.Writer, s Struct) {
	// unmarshal code
	fmt.Fprintf(w, `
func (strct *%s) UnmarshalJSON(b []byte) error {
    var jsonMap map[string]json.RawMessage
    if err := json.Unmarshal(b, &jsonMap); err != nil {
        return err
    }
    // first parse all the defined properties
    for k, v := range jsonMap {
        switch k {
`, s.Name)
	// handle defined properties
	for _, fieldKey := range getOrderedFieldNames(s.Fields) {
		f := s.Fields[fieldKey]
		if f.JSONName == "-" {
			continue
		}
		fmt.Fprintf(w, `        case "%s":
            if err := json.Unmarshal([]byte(v), &strct.%s); err != nil {
                return err
             }
`, f.JSONName, f.Name)
	}
	// now handle additional values
	initialiser, isPrimitive := getPrimitiveInitialiser(s.AdditionalValueType)
	addressOfInitialiser := "&"
	if isPrimitive {
		addressOfInitialiser = ""
	}
	fmt.Fprintf(w, `        default:
            // an additional "%s" value
            additionalValue := %s
            if err := json.Unmarshal([]byte(v), &additionalValue); err != nil {
                return err
            }
            if strct.AdditionalProperties == nil {
                strct.AdditionalProperties = make(map[string]%s, 0)
            }
            strct.AdditionalProperties[k]= %sadditionalValue
`, s.AdditionalValueType, initialiser, s.AdditionalValueType, addressOfInitialiser)
	fmt.Fprintf(w, "        }\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "    return nil\n")
	fmt.Fprintf(w, "}\n")
}

func getPrimitiveInitialiser(typ string) (string, bool) {
	// strip *pointer dereference symbol so we can use in declaration
	deref := strings.Replace(typ, "*", "", 1)
	switch {
	case strings.HasPrefix(deref, "int"):
		return "0", true
	case strings.HasPrefix(deref, "float"):
		return "0.0", true
	case deref == "string":
		return "\"\"", true
	}
	return deref + "{}", false
}

func outputNameAndDescriptionComment(name, description string, w io.Writer) {
	if strings.Index(description, "\n") == -1 {
		fmt.Fprintf(w, "// %s %s\n", name, description)
		return
	}

	dl := strings.Split(description, "\n")
	fmt.Fprintf(w, "// %s %s\n", name, strings.Join(dl, "\n// "))
}

