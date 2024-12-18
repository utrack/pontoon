// Package docgen provides tools for extracting and storing documentation from Go source code.
package docgen

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// DocID uniquely identifies a documentation entity in the codebase.
type DocID struct {
	PkgPath    string   // Full package import path
	TypeName   string   // Type name (for services/structs)
	MethodName string   // Method name (for handlers)
	FieldPath  []string // Field path for nested structs
	FilePath   string   // File path relative to module root
	Line       int      // Line number in source
}

// Hash generates a stable hash for runtime lookup.
func (id DocID) Hash() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s/%s.%s/%s:%d",
		id.PkgPath,
		id.TypeName,
		id.MethodName,
		strings.Join(id.FieldPath, "."),
		id.Line,
	)))
	return fmt.Sprintf("%x", h.Sum(nil)[:8]) // First 8 bytes is enough
}

// DocComment represents a parsed documentation comment.
type DocComment struct {
	ID         string // Stable hash for runtime lookup
	Path       string // Full path to the type/field
	Comment    string // Raw comment text
	SourceFile string // Source file path
	Line       int    // Line number in source
	Type       string // Comment type: "service", "handler", "struct", "field"
	Identifier DocID  // Full identifier information
}

// ServiceDoc represents documentation for a service.
type ServiceDoc struct {
	Name     string
	Package  string
	File     string // Source file path
	Line     int    // Line number in source
	Comments []DocComment
	Methods  []MethodDoc
	Types    map[string]TypeDoc // All referenced types
}

// MethodDoc represents documentation for a service method.
type MethodDoc struct {
	Name       string
	Comment    string
	File       string
	Line       int
	InputType  string
	OutputType string
	ReturnsWellFormedError bool
}

// TypeDoc represents documentation for a type.
type TypeDoc struct {
	Name     string
	Package  string
	File     string
	Line     int
	Comment  string
	Fields   []FieldDoc
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
