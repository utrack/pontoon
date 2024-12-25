// Package docgen provides tools for extracting and storing documentation from Go source code.
package docgen

import (
)


// FunctionDoc is some function's documentation.
type FunctionDoc struct {
	Package string
	Name    string
	Comment string
	File    string
	Line    int
	Params  []FunctionParamDoc
	Returns []FunctionParamDoc
}

type FunctionParamDoc struct {
	Name string
	Type string
}

// TypeDoc represents documentation for a type.
type TypeDoc struct {
	Name     string
	Package  string
	File     string
	Line     int
	Comment  string
	Fields   []FieldDoc
	Methods  []FunctionDoc
	IsStruct bool
}

// FieldDoc represents documentation for a struct field.
type FieldDoc struct {
	Name       string
	Type       string // Full type name
	Comment    string
	Tags       string
	FilePath   string
	Line       int
	Nullable   bool
	IsEmbedded bool
	IsMap      *FieldMapDoc
	IsArray    *FieldArrayDoc
}

type FieldArrayDoc struct {
	Type string
}

type FieldMapDoc struct {
	TypeKey   string
	TypeValue string
}
