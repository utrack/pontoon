package main

import (
	"encoding/json"
	"go/importer"
	"testing"
)

func TestRawMessageFieldSchemaIsUnconstrained(t *testing.T) {
	jsonPackage, err := importer.Default().Import("encoding/json")
	if err != nil {
		t.Fatal(err)
	}

	rawMessage := jsonPackage.Scope().Lookup("RawMessage")
	if rawMessage == nil {
		t.Fatal("encoding/json.RawMessage type not found")
	}

	typeDescription, err := (&builder{}).getTypeDesc(rawMessage.Type())
	if err != nil {
		t.Fatal(err)
	}

	schema, err := genFieldSchema(descField{t: typeDescription})
	if err != nil {
		t.Fatal(err)
	}
	actual, err := json.Marshal(schema.Value)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != "{}" {
		t.Fatalf("json.RawMessage schema = %s, want {}", actual)
	}
}
