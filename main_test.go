package main

import (
	"encoding/json"
	"testing"
)

func TestDTOJSONMatchesJavaRecord(t *testing.T) {
	value, err := json.Marshal(dto{I: 10, S: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != `{"i":10,"s":"test"}` {
		t.Fatalf("unexpected JSON: %s", value)
	}
}
