package generate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

// ReadInputFiles from disk and convert to JSON schema.
func ReadInputFiles(inputFiles []string, schemaKeyRequired bool) ([]*Schema, error) {
	schemas := make([]*Schema, len(inputFiles))
	for i, file := range inputFiles {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, errors.New("failed to read the input file with error " + err.Error())
		}

		abPath, err := abs(file)
		if err != nil {
			return nil, errors.New("failed to normalise input path with error " + err.Error())
		}

		fileURI := url.URL{
			Scheme: "file",
			Path:   abPath,
		}

		schemas[i], err = ParseWithSchemaKeyRequired(string(b), &fileURI, schemaKeyRequired)
		if err != nil {
			if jsonError, ok := err.(*json.SyntaxError); ok {
				line, character, lcErr := lineAndCharacter(b, int(jsonError.Offset))
				errStr := fmt.Sprintf("cannot parse JSON schema due to a syntax error at %s line %d, character %d: %v\n", file, line, character, jsonError.Error())
				if lcErr != nil {
					errStr += fmt.Sprintf("couldn't find the line and character position of the error due to error %v\n", lcErr)
				}
				return nil, errors.New(errStr)
			}
			if jsonError, ok := err.(*json.UnmarshalTypeError); ok {
				line, character, lcErr := lineAndCharacter(b, int(jsonError.Offset))
				errStr := fmt.Sprintf("the JSON type '%v' cannot be converted into the Go '%v' type on struct '%s', field '%v'. See input file %s line %d, character %d\n", jsonError.Value, jsonError.Type.Name(), jsonError.Struct, jsonError.Field, file, line, character)
				if lcErr != nil {
					errStr += fmt.Sprintf("couldn't find the line and character position of the error due to error %v\n", lcErr)
				}
				return nil, errors.New(errStr)
			}
			return nil, fmt.Errorf("failed to parse the input JSON schema file %s with error %v", file, err)
		}
	}

	return schemas, nil
}

func lineAndCharacter(bytes []byte, offset int) (line int, character int, err error) {
	lf := byte(0x0A)

	if offset > len(bytes) {
		return 0, 0, fmt.Errorf("couldn't find offset %d in %d bytes", offset, len(bytes))
	}

	// Humans tend to count from 1.
	line = 1

	for i, b := range bytes {
		if b == lf {
			line++
			character = 0
		}
		character++
		if i == offset {
			return line, character, nil
		}
	}

	return 0, 0, fmt.Errorf("couldn't find offset %d in %d bytes", offset, len(bytes))
}

func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
