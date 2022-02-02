package test

import (
	"encoding/json"
	"testing"

	gen "github.com/a-h/generate/test/oneofnull_gen"
)

func TestOneOfNullParsingValue(t *testing.T) {

	noPropData := `{"optional_name": "v"}`

	s := gen.Simple{}
	err := json.Unmarshal([]byte(noPropData), &s)
	if err != nil {
		t.Fatal(err)
	}

	if s.OptionalName == nil {
		t.Fatal("Field option_name should be initialized")
	}

	if *s.OptionalName != "v" {
		t.Fatal("Field option_name should be `v`")
	}

}

func TestOneOfNullParsingNull(t *testing.T) {

	noPropData := `{"optional_name": null}`

	s := gen.Simple{}
	err := json.Unmarshal([]byte(noPropData), &s)
	if err != nil {
		t.Fatal(err)
	}

	if s.OptionalName != nil {
		t.Fatal("Field option_name should be nil")
	}

}

func TestOneOfNullParsingAbsent(t *testing.T) {

	noPropData := `{}`

	s := gen.Simple{}
	err := json.Unmarshal([]byte(noPropData), &s)
	if err != nil {
		t.Fatal(err)
	}

	if s.OptionalName != nil {
		t.Fatal("Field option_name should be nil")
	}

}
