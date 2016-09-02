package main

import (
	"encoding/json"
	"testing"
)

func TestThatJSONCanBeRoundtrippedUsingGeneratedStructs(t *testing.T) {
	j := `{"address":{"county":"countyValue"},"name":"nameValue"}`

	e := &Example{}
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

type Address struct {
	County      string `json:"county,omitempty"`
	District    string `json:"district,omitempty"`
	FlatNumber  string `json:"flatNumber,omitempty"`
	HouseName   string `json:"houseName,omitempty"`
	HouseNumber string `json:"houseNumber,omitempty"`
	Postcode    string `json:"postcode,omitempty"`
	Street      string `json:"street,omitempty"`
	Town        string `json:"town,omitempty"`
}

type Example struct {
	Address *Address `json:"address,omitempty"`
	Name    string   `json:"name,omitempty"`
	Status  *Status  `json:"status,omitempty"`
}

type Status struct {
	Favouritecat string `json:"favouritecat,omitempty"`
}
