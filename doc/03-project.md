# Project Outline

Pontoon is an OpenAPI 3.1-compilant API documentation generator.
It describes the Go structures and HTTP handlers in OpenAPI format according to how they're parsed by `github.com/ggicci/httpin`.

To find the handler and structure, it compares the method to the `sdesc.Service` interface signature.

It extracts the path, HTTP verb, function name and in/out parameters by parsing the implementation's `RegisterHTTP` method.

When generating, it preserves Go comments, translating them into OpenAPI description documentation.

## Traits

- AST-parses the Go files to extract the comments and documentation for the handler's structs.
- TODO: Generate OpenAPI documentation in runtime and merge it with the comments

## Documentation Generation

The `docgen` tool extracts documentation from Go source code comments and generates YAML documentation files.

Documentation is embedded in Go source files for runtime access, with the following structure:

```yaml
services:
  ServiceName:
    id: "package/path.ServiceName"
    type: "service"
    methods:
      methodName:
        input_type: "package/path.InputType"
        output_type: "package/path.OutputType"
        
types:
  "package/path.TypeName":
    package: "package/path"
    name: "TypeName"
    fields:
      - name: "FieldName"
        type: "string"
        nullable: false
        tags: 'json:"field_name"'
```
