package test

import (
	"testing"

	defaultValues "github.com/a-h/generate/test/defaultValues_gen"
)

func TestDefaultValues(t *testing.T) {
	result := &defaultValues.Root{}
	result.UnmarshalJSON([]byte("{}"))
	if result.Name != "Unnamed" {
		t.Errorf("Object had unexpected name \"%s\".", result.Name)
	}
	if result.Age != -1 {
		t.Errorf("Object had unexpected name \"%d\".", result.Age)
	}
	if result.Score != -1.0 {
		t.Errorf("Object had unexpected score \"%f\".", result.Score)
	}
}
