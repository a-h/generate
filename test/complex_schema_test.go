package test

import (
	"github.com/azarc-io/json-schema-to-go-struct-generator/pkg/converter"
	"os"
	"path/filepath"
	"testing"
)

//go:generate go run ../cmd/main.go --input ./samples/complex-schemas --output ./generated/complex-schemas

func TestConvert(t *testing.T) {
	wd, _ := os.Getwd()
	filePath := filepath.Join(wd, "samples/complex-schemas/schema_cusdec-ifd-ins_full-vth-schema.json")
	err := converter.Convert([]string{filePath}, "./generated/complex-schemas")
	if err != nil {
		t.Error(err)
	}
}
