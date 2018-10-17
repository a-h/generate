package generate

import (
	"reflect"
	"strings"
	"testing"
)

func TestThatFieldNamesAreOrdered(t *testing.T) {
	m := map[string]Field{
		"z": {},
		"b": {},
	}

	actual := getOrderedFieldNames(m)
	expected := []string{"b", "z"}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %s and actual %s should match in order", strings.Join(expected, ", "), strings.Join(actual, ","))
	}
}

func TestThatStructNamesAreOrdered(t *testing.T) {
	m := map[string]Struct{
		"c": {},
		"b": {},
		"a": {},
	}

	actual := getOrderedStructNames(m)
	expected := []string{"a", "b", "c"}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %s and actual %s should match in order", strings.Join(expected, ", "), strings.Join(actual, ","))
	}
}

func TestLineAndCharacterFromOffset(t *testing.T) {
	tests := []struct {
		In                []byte
		Offset            int
		ExpectedLine      int
		ExpectedCharacter int
		ExpectedError     bool
	}{
		{
			In:                []byte("Line 1\nLine 2"),
			Offset:            6,
			ExpectedLine:      2,
			ExpectedCharacter: 1,
		},
		{
			In:                []byte("Line 1\r\nLine 2"),
			Offset:            7,
			ExpectedLine:      2,
			ExpectedCharacter: 1,
		},
		{
			In:                []byte("Line 1\nLine 2"),
			Offset:            0,
			ExpectedLine:      1,
			ExpectedCharacter: 1,
		},
		{
			In:                []byte("Line 1\nLine 2"),
			Offset:            200,
			ExpectedLine:      0,
			ExpectedCharacter: 0,
			ExpectedError:     true,
		},
		{
			In:                []byte("Line 1\nLine 2"),
			Offset:            -1,
			ExpectedLine:      0,
			ExpectedCharacter: 0,
			ExpectedError:     true,
		},
	}

	for _, test := range tests {
		actualLine, actualCharacter, err := lineAndCharacter(test.In, test.Offset)
		if err != nil && !test.ExpectedError {
			t.Errorf("Unexpected error for input %s at offset %d: %v", test.In, test.Offset, err)
			continue
		}

		if actualLine != test.ExpectedLine || actualCharacter != test.ExpectedCharacter {
			t.Errorf("For '%s' at offset %d, expected %d:%d, but got %d:%d", test.In, test.Offset, test.ExpectedLine, test.ExpectedCharacter, actualLine, actualCharacter)
		}
	}
}
