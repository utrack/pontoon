# Project Outline

Pontoon is an OpenAPI 3.1-compilant API documentation generator.
It describes the Go structures and HTTP handlers in OpenAPI format according to how they're parsed by `github.com/ggicci/httpin`.

To find the handler and structure, it compares the method to the `sdesc.Service` interface signature.

It extracts the path, HTTP verb, function name and in/out parameters by parsing the implementation's `RegisterHTTP` method.

When generating, it preserves Go comments, translating them into OpenAPI description documentation.

## Traits

- AST-parses the Go files before generating the OpenAPI spec to a separate <filename>.pontoon.go file.
- No actions on runtime
