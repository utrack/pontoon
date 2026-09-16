package main

import (
	"encoding/json"
	"go/importer"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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

func TestJSONTextValueFieldSchemaIsUnconstrained(t *testing.T) {
	if !supportsJSONV2Embed {
		t.Skip("requires Go 1.27")
	}
	jsonTextPackage, err := importer.Default().Import("encoding/json/jsontext")
	if err != nil {
		t.Fatal(err)
	}

	value := jsonTextPackage.Scope().Lookup("Value")
	if value == nil {
		t.Fatal("encoding/json/jsontext.Value type not found")
	}

	typeDescription, err := (&builder{}).getTypeDesc(value.Type())
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
		t.Fatalf("jsontext.Value schema = %s, want {}", actual)
	}
}

func TestJSONFieldTag(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		wantName    string
		wantHasName bool
		wantEmbed   bool
	}{
		{name: "Value", tag: `json:",embed"`, wantName: "Value", wantEmbed: true},
		{name: "Value", tag: `json:"value,omitempty"`, wantName: "value", wantHasName: true},
		{name: "Value", tag: `json:"value,omitzero"`, wantName: "value", wantHasName: true},
		{name: "Value", tag: `json:"value,string,case:strict"`, wantName: "value", wantHasName: true},
	}

	for _, test := range tests {
		got := parseJSONFieldTag(test.name, test.tag)
		if got.name != test.wantName || got.hasName != test.wantHasName || got.embed != test.wantEmbed {
			t.Errorf("parseJSONFieldTag(%q, %q) = %#v", test.name, test.tag, got)
		}
	}
}

func TestJSONEmbeddedField(t *testing.T) {
	tests := []struct {
		goEmbedded bool
		tag        jsonFieldTag
		want       bool
	}{
		{goEmbedded: true, tag: parseJSONFieldTag("Embedded", ""), want: true},
		{goEmbedded: true, tag: parseJSONFieldTag("Embedded", `json:"named"`)},
		{tag: parseJSONFieldTag("Embedded", `json:",embed"`), want: supportsJSONV2Embed},
		{tag: parseJSONFieldTag("Embedded", "")},
	}

	for _, test := range tests {
		if got := isJSONEmbeddedField(test.goEmbedded, test.tag); got != test.want {
			t.Errorf("isJSONEmbeddedField(%v, %#v) = %v, want %v", test.goEmbedded, test.tag, got, test.want)
		}
	}
}

func TestStructFieldEmbedUsesComposition(t *testing.T) {
	cacheSchemaRefs = map[*typeDesc]*openapi3.SchemaRef{}
	embedded := &typeDesc{
		typeName: "test.Embedded",
		isStruct: &descStruct{},
	}
	parent := &typeDesc{
		typeName: "test.Parent",
		isStruct: &descStruct{embeds: []descField{{
			name: "Embedded",
			tags: `json:",embed"`,
			t:    embedded,
		}}},
	}

	schema, err := genRefFieldStruct(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Value.AllOf) != 1 || schema.Value.AllOf[0].Ref != "#/components/schemas/test.Embedded" {
		t.Fatalf("embedded struct allOf = %#v", schema.Value.AllOf)
	}
}

func TestMapFieldEmbedUsesTypedAdditionalProperties(t *testing.T) {
	if !supportsJSONV2Embed {
		t.Skip("requires Go 1.27")
	}
	cacheSchemaRefs = map[*typeDesc]*openapi3.SchemaRef{}
	parent := &typeDesc{
		typeName: "test.Parent",
		isStruct: &descStruct{embeds: []descField{{
			name: "Values",
			tags: `json:",embed"`,
			t: &typeDesc{isMap: &descMap{
				key:   &typeDesc{isScalar: true, typeName: "string"},
				value: &typeDesc{isScalar: true, typeName: "int"},
			}},
		}}},
	}

	schema, err := genRefFieldStruct(parent)
	if err != nil {
		t.Fatal(err)
	}
	additional := schema.Value.AdditionalProperties
	if additional == nil || additional.Value == nil || additional.Value.Type != "integer" {
		t.Fatalf("embedded map additionalProperties = %#v", additional)
	}
}

func TestJSONTextValueEmbedAllowsAnyAdditionalProperties(t *testing.T) {
	if !supportsJSONV2Embed {
		t.Skip("requires Go 1.27")
	}
	cacheSchemaRefs = map[*typeDesc]*openapi3.SchemaRef{}
	parent := &typeDesc{
		typeName: "test.Parent",
		isStruct: &descStruct{embeds: []descField{{
			name: "Value",
			tags: `json:",embed"`,
			t:    &typeDesc{id: "encoding/json/jsontext.Value", isAny: true},
		}}},
	}

	schema, err := genRefFieldStruct(parent)
	if err != nil {
		t.Fatal(err)
	}
	allowed := schema.Value.AdditionalPropertiesAllowed
	if allowed == nil || !*allowed {
		t.Fatalf("embedded jsontext.Value additionalPropertiesAllowed = %v", allowed)
	}
	if _, exists := schema.Value.Properties[",embed"]; exists {
		t.Fatal("embedded jsontext.Value generated a literal ,embed property")
	}
}
