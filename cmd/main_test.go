package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/a-h/generate"
)

func TestThatFieldNamesAreOrdered(t *testing.T) {
	m := map[string]generate.Field{
		"z": generate.Field{},
		"b": generate.Field{},
	}

	actual := getOrderedFieldNames(m)
	expected := []string{"b", "z"}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %s and actual %s should match in order", strings.Join(expected, ", "), strings.Join(actual, ","))
	}
}

func TestThatStructNamesAreOrdered(t *testing.T) {
	m := map[string]generate.Struct{
		"c": generate.Struct{},
		"b": generate.Struct{},
		"a": generate.Struct{},
	}

	actual := getOrderedStructNames(m)
	expected := []string{"a", "b", "c"}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %s and actual %s should match in order", strings.Join(expected, ", "), strings.Join(actual, ","))
	}
}
