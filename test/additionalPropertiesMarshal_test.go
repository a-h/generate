package test

import (
	"encoding/json"
	gen "github.com/giorgos-nikolopoulos/generate/test/additionalPropertiesMarshal_gen"
	"reflect"
	"testing"
)

func TestApRefNoProp(t *testing.T) {
}

func TestApRefProp(t *testing.T) {
}

func TestApRefReqProp(t *testing.T) {
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestApTrueNoProp(t *testing.T) {

	noPropData := `{"a": "b", "c": 42 }`

	ap := gen.ApTrueNoProp{}
	err := json.Unmarshal([]byte(noPropData), &ap)
	if err != nil {
		t.Fatal(err)
	}

	if len(ap.AdditionalProperties) != 2 {
		t.Fatalf("Wrong number of additionalProperties: %d", len(ap.AdditionalProperties))
	}

	if s, ok := ap.AdditionalProperties["a"].(string); !ok {
		t.Fatalf("a was not a string")
	} else {
		if s != "b" {
			t.Fatalf("wrong value for a: \"%s\" (should be \"b\")", s)
		}
	}

	typ := reflect.TypeOf(ap.AdditionalProperties["c"])

	if c, ok := ap.AdditionalProperties["c"].(float64); !ok {
		t.Fatalf("c was not an number (it was a %s)", typ.Name())
	} else {
		if c != 42 {
			t.Fatalf("wrong value for c: \"%f\" (should be \"42\")", c)
		}
	}
}

func TestApTrueProp(t *testing.T) {
	data := `{"a": "b", "c": 42, "stuff": "xyz" }`

	ap := gen.ApTrueProp{}
	err := json.Unmarshal([]byte(data), &ap)
	if err != nil {
		t.Fatal(err)
	}

	if len(ap.AdditionalProperties) != 2 {
		t.Fatalf("Wrong number of additionalProperties: %d", len(ap.AdditionalProperties))
	}

	if s, ok := ap.AdditionalProperties["a"].(string); !ok {
		t.Fatalf("a was not a string")
	} else {
		if s != "b" {
			t.Fatalf("wrong value for a: \"%s\" (should be \"b\")", s)
		}
	}

	typ := reflect.TypeOf(ap.AdditionalProperties["c"])

	if c, ok := ap.AdditionalProperties["c"].(float64); !ok {
		t.Fatalf("c was not an number (it was a %s)", typ.Name())
	} else {
		if c != 42 {
			t.Fatalf("wrong value for c: \"%f\" (should be \"42\")", c)
		}
	}

	if ap.Stuff != "xyz" {
		t.Fatalf("invalid stuff value: \"%s\"", ap.Stuff)
	}
}

func TestApTrueReqProp(t *testing.T) {
	dataGood := `{"a": "b", "c": 42, "stuff": "xyz" }`
	dataBad := `{"a": "b", "c": 42 }`

	{
		ap := gen.ApTrueReqProp{}
		err := json.Unmarshal([]byte(dataBad), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	ap := gen.ApTrueReqProp{}
	err := json.Unmarshal([]byte(dataGood), &ap)
	if err != nil {
		t.Fatal(err)
	}

	if len(ap.AdditionalProperties) != 2 {
		t.Fatalf("Wrong number of additionalProperties: %d", len(ap.AdditionalProperties))
	}

	if s, ok := ap.AdditionalProperties["a"].(string); !ok {
		t.Fatalf("a was not a string")
	} else {
		if s != "b" {
			t.Fatalf("wrong value for a: \"%s\" (should be \"b\")", s)
		}
	}

	typ := reflect.TypeOf(ap.AdditionalProperties["c"])

	if c, ok := ap.AdditionalProperties["c"].(float64); !ok {
		t.Fatalf("c was not an number (it was a %s)", typ.Name())
	} else {
		if c != 42 {
			t.Fatalf("wrong value for c: \"%f\" (should be \"42\")", c)
		}
	}

	if ap.Stuff != "xyz" {
		t.Fatalf("invalid stuff value: \"%s\"", ap.Stuff)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestApFalseNoProp(t *testing.T) {

	dataBad1 := `{"a": "b", "c": 42, "stuff": "xyz"}`
	dataBad2 := `{"a": "b", "c": 42}`
	dataBad3 := `{"stuff": "xyz"}`
	dataGood1 := `{}`

	{
		ap := gen.ApFalseNoProp{}
		err := json.Unmarshal([]byte(dataBad1), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseNoProp{}
		err := json.Unmarshal([]byte(dataBad2), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseNoProp{}
		err := json.Unmarshal([]byte(dataBad3), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseNoProp{}
		err := json.Unmarshal([]byte(dataGood1), &ap)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestApFalseProp(t *testing.T) {
	dataBad1 := `{"a": "b", "c": 42, "stuff": "xyz"}`
	dataBad2 := `{"a": "b", "c": 42}`
	dataGood1 := `{"stuff": "xyz"}`
	dataGood2 := `{}`

	{
		ap := gen.ApFalseProp{}
		err := json.Unmarshal([]byte(dataBad1), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseProp{}
		err := json.Unmarshal([]byte(dataBad2), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseProp{}
		err := json.Unmarshal([]byte(dataGood1), &ap)
		if err != nil {
			t.Fatal(err)
		}
		if ap.Stuff != "xyz" {
			t.Fatalf("invalid stuff value: \"%s\"", ap.Stuff)
		}
	}

	{
		ap := gen.ApFalseProp{}
		err := json.Unmarshal([]byte(dataGood2), &ap)
		if err != nil {
			t.Fatal(err)
		}
		if ap.Stuff != "" {
			t.Fatalf("invalid stuff value: \"%s\"", ap.Stuff)
		}
	}

	ap := gen.ApFalseProp{}
	if _, ok := reflect.TypeOf(ap).FieldByName("AdditionalProperties"); ok {
		t.Fatalf("AdditionalProperties was generated where it should not have been")
	}
}

func TestApFalseReqProp(t *testing.T) {
	dataBad1 := `{"a": "b", "c": 42, "stuff": "xyz"}`
	dataBad2 := `{"a": "b", "c": 42}`
	dataBad3 := `{}`
	dataGood := `{"stuff": "xyz"}`

	{
		ap := gen.ApFalseReqProp{}
		err := json.Unmarshal([]byte(dataBad1), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseReqProp{}
		err := json.Unmarshal([]byte(dataBad2), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	{
		ap := gen.ApFalseReqProp{}
		err := json.Unmarshal([]byte(dataBad3), &ap)
		if err == nil {
			t.Fatalf("should have returned an error, required field missing")
		}
	}

	ap := gen.ApFalseReqProp{}
	err := json.Unmarshal([]byte(dataGood), &ap)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := reflect.TypeOf(ap).FieldByName("AdditionalProperties"); ok {
		t.Fatalf("AdditionalProperties was generated where it should not have been")
	}

	if ap.Stuff != "xyz" {
		t.Fatalf("invalid stuff value: \"%s\"", ap.Stuff)
	}
}
