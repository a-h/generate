package main

import (
	"bytes"
	"io"
	"io/ioutil"
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

func TestThatThePackageCanBeSet(t *testing.T) {
	pkg := "testpackage"
	p = &pkg

	r, w := io.Pipe()

	go Output(w, make(map[string]generate.Struct))

	lr := io.LimitedReader{R: r, N: 20}
	bs, _ := ioutil.ReadAll(&lr)
	output := bytes.NewBuffer(bs).String()

	if output != "package testpackage\n" {
		t.Error("Unexpected package declaration: ", output)
	}
}
