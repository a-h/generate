package test

import (
	"encoding/json"
	gen "github.com/azarc-io/json-schema-to-go-struct-generator/test/generated/marshal/test"
	"testing"
)

//go:generate go run ../cmd/main.go --input ./samples/marshal --output ./generated/marshal

func TestThatJSONCanBeRoundtrippedUsingGeneratedStructs(t *testing.T) {
	j := `{"address":{"county":"countyValue"},"name":"nameValue"}`

	e := &gen.Example{}
	err := json.Unmarshal([]byte(j), e)

	if err != nil {
		t.Fatal("Failed to unmarshall JSON with error ", err)
	}

	if e.Address.County != "countyValue" {
		t.Errorf("the county value was not found, expected 'countyValue' got '%s'", e.Address.County)
	}

	op, err := json.Marshal(e)

	if err != nil {
		t.Error("Failed to marshal JSON with error ", err)
	}

	if string(op) != j {
		t.Errorf("expected %s, got %s", j, string(op))
	}
}
