package test

import (
	"strings"
	"encoding/json"
	"github.com/a-h/generate/test/example1_gen"
	"reflect"
	"testing"
)

func TestExample1(t *testing.T) {
	params := []struct {
		Name           string
		Data           string
		ExpectedResult bool
	}{
		{
			Name: "Blue Sky",
			Data: `{
				"id": 1,
				"name": "Unbridled Optimism 2.0",
				"price": 99.99,
				"tags": [ "happy" ] }`,
			ExpectedResult: true,
		},
		{
			Name: "Missing Price",
			Data: `{
				"id": 1,
				"name": "Unbridled Optimism 2.0",
				"tags": [ "happy" ] }`,
			ExpectedResult: false,
		},
	}

	for _, param := range params {

		prod := &example1.Product{}
		if err := json.Unmarshal([]byte(param.Data), &prod); err != nil {
			if param.ExpectedResult {
				t.Fatal(err)
			}
		} else {
			if !param.ExpectedResult {
				t.Fatal("Expected failure, got success: " + param.Name)
			}
		}
	}
}

func TestExample1Access(t *testing.T) {
	fs := reflect.VisibleFields(reflect.TypeOf(example1.Product{}))
	for _, f := range fs {
		if f.Name == "CouponCode" {
			if v, ok := f.Tag.Lookup("jsonSchema"); !ok {
				t.Fatal("Expected CouponCode field to have jsonSchema tag")
			} else if !strings.Contains(v, "writeonly") {
				t.Fatalf("CouponCode's json %q doesn't specify writeonly", v)
			}
		}

		if f.Name == "InStock" {
			if v, ok := f.Tag.Lookup("jsonSchema"); !ok {
				t.Fatal("Expected InStock field to have jsonSchema tag")
			} else if !strings.Contains(v, "readonly") {
				t.Fatalf("InStock's json %q doesn't specify readonly", v)
			}
		}
	}
}
