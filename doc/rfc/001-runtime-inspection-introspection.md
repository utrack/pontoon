# Runtime Struct Introspection

> Part of [Runtime Inspection](/doc/features/001-runtime-inspection.md) feature.

## Overview
This RFC describes the implementation of runtime struct introspection for httpin-annotated structs, generating OpenAPI 3.1 Component Schemas. This is the second part of the Runtime Inspection feature, focusing on runtime model generation.

## Goals
- Generate OpenAPI 3.1 Component Schemas from httpin-annotated structs at runtime
- Support standalone usage and Pontoon integration
- Handle all non-validation httpin annotations
- Support OpenAPI 3.1 nullability semantics

## Non-Goals
- Generating any other OpenAPI components (paths, security, etc.)
- Supporting validation annotations
- Supporting other annotation systems besides httpin
- Generating documentation from comments (handled by precompile step)

## Implementation

### Package Structure
```
pontoon/
└── openapi/
    └── httpinmeditate/
        ├── httpin.go     # httpin-specific reflection code
        ├── schema.go     # OpenAPI schema generation
        └── convert.go    # type conversion utilities
```

### Core Components

#### 1. Type Reflection (httpin.go)
- Create a type walker that traverses struct fields
- Extract httpin field tags
- Build an intermediate representation (IR) of the type structure
- Handle nested structs and complex types (arrays, maps, etc.)

#### 2. Schema Generation (schema.go)
- Convert IR to OpenAPI 3.1 Component Schema using libopenapi
- Support all OpenAPI 3.1 primitive types
- Handle nullability according to OpenAPI 3.1 spec
- Generate proper references for nested types

#### 3. Type Conversion (convert.go)
- Map Go primitive types to OpenAPI types
- Handle special cases (time.Time, encoding.TextMarshaler, etc.)
- Support custom primitive type conversions

### Type Handling Rules

#### Go Types
- Interface types are not supported and will return an error
- Embedded structs are declared as separate schemas and merged via OpenAPI `allOf`
- Pointer types are marked as nullable using OpenAPI 3.1 `oneOf` with `null`
- All nested types are referenced via `$ref` for better maintainability
- Unexported fields are ignored

Example of pointer type schema:
```yaml
# *string
oneOf:
  - type: 'null'
  - type: string

# *UserProfile
oneOf:
  - type: 'null'
  - $ref: '#/components/schemas/UserProfile'
```

Example of embedded struct:
```yaml
# type Request struct {
#   CommonFields
#   name string
# }
allOf:
  - $ref: '#/components/schemas/CommonFields'
  - type: object
    properties:
      name:
        type: string
```

#### httpin Annotations
Supported annotations:
- `form` - form-data bindings
- `query` - URL query parameters
- `header` - HTTP headers
- `cookie` - HTTP cookies
- `path` - URL path parameters
- `json` - JSON body fields

Validation annotations (not supported):
- `required` - field requirements
- `default` - default values
- `enum` - enumeration constraints
- `pattern` - regex patterns
- `range` - numeric ranges

Default values from httpin tags are converted to OpenAPI default values.

### Public API

```go
// Package httpinmeditate provides OpenAPI 3.1 schema generation for httpin-annotated structs
package httpinmeditate

// Generator generates OpenAPI 3.1 Component Schemas from httpin-annotated structs
type Generator struct {
    // SchemaNameFunc allows customizing schema names for types
    SchemaNameFunc func(reflect.Type) string
}

// NewGenerator creates a new schema generator with default options
func NewGenerator() *Generator

// GenerateSchema generates an OpenAPI 3.1 Component Schema for the given type
func (g *Generator) GenerateSchema(t reflect.Type) (*openapi3.Schema, error)

// GenerateSchemaWithName generates a named schema and adds it to Components
func (g *Generator) GenerateSchemaWithName(name string, t reflect.Type) (*openapi3.Components, error)
```

### Error Cases
- Interface types in struct fields
- Circular type dependencies
- Missing or malformed httpin tags
- Unsupported Go types (channels, functions)

### Integration with Pontoon

The package will be used in two ways:
1. Standalone - users can generate schemas for any httpin-annotated structs
2. Integrated - Pontoon will use it as part of its OpenAPI generation pipeline

### Error Handling
- Return descriptive errors for unsupported types
- Validate input types are struct types
- Check for httpin tag presence
- Report circular dependencies

### Dependencies
- github.com/pb33f/libopenapi v0.7.0+ for OpenAPI model creation
- github.com/ggicci/httpin for annotation parsing

## Testing Strategy
1. Unit tests:
   - Test each primitive type conversion
   - Verify schema generation for basic types
   - Test embedded struct handling and allOf generation
   - Verify pointer type schema generation with oneOf
   - Test proper $ref generation for nested types
   - Verify httpin tag parsing and default value conversion
   - Check error cases
2. Integration tests:
   - Test with real httpin-annotated structs
   - Verify schema compatibility with OpenAPI tools
3. Benchmarks:
   - Measure performance on large structs
   - Check memory allocations

## Future Considerations
1. Cache generated schemas for better performance
2. Add support for validation annotations
3. Consider supporting custom struct handling strategies (inline vs reference)

## Migration Strategy
No migration needed as this is a new feature.

## Security Considerations
- Ensure reflection is safe and doesn't expose sensitive fields
- Validate all user input before processing
- Consider resource limits for large types

## Related Documents
- [Runtime Inspection Feature](/doc/features/001-runtime-inspection.md)
- [Runtime Inspection Precompile RFC](/doc/rfc/001-runtime-inspection-precompile.md)
