package test

import (
	"testing"

	abandoned "github.com/azarc-io/json-schema-to-go-struct-generator/test/generated/abandoned/abandoned"
)

//go:generate go run ../cmd/main.go --input ./samples/abandoned --output ./generated/abandoned

func TestAbandoned(t *testing.T) {
	// this just tests the name generation works correctly
	r := abandoned.Root{
		Name:      "jonson",
		Abandoned: &abandoned.PackageList{},
	}
	// the test is the presence of the Abandoned field
	if r.Abandoned == nil {
		t.Fatal("thats the test")
	}
}
