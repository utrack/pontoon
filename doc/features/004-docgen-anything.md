# Feature: Docgen Anything

## Overview
Currently, Pontoon's docgen tool only generates documentation for handlers that implement the `Service` interface. This feature adds an optional `--all` flag to generate documentation for any package, enabling complete documentation for packages like `httpinoapi`. Without the flag, docgen maintains its default behavior of documenting only Service handlers.

## User Stories
1. As a developer, I want to generate documentation for any Go package using the `--all` flag
2. As a developer, I want the default behavior (without `--all`) to remain focused on Service handlers
3. As a developer using `httpinoapi`, I want to have complete package documentation
4. As a developer, I want warnings (not errors) when types can't be fully documented, with partial documentation still generated

## Technical Constraints
- Must maintain compatibility with existing docgen functionality
- Should reuse existing documentation extraction and generation logic where possible
- Must preserve Go comments translation into OpenAPI documentation
- Must support recursive struct types
- Must convert errors to warnings in --all mode to ensure partial documentation is always generated

## Integration Points
- Integrates with existing `docgen` package
- Affects `httpinoapi` package documentation generation
- Interfaces with AST parsing for type discovery
- Uses existing YAML documentation format
- Maintains OpenAPI documentation generation pipeline

## Corner Cases
1. Package contains both Service and non-Service types
2. Package contains types that are used by multiple Services
3. Package contains types with complex relationships (embedded structs, interfaces)
4. Package documentation might overlap with Service documentation
5. Types might have dependencies in other packages not covered by --all
6. Package might contain generated code that should be excluded
7. Recursive struct types must be properly handled

## Implementation Plan (High Level)
1. Add --all flag to docgen CLI
   - Add flag definition and parsing
   - Pass flag value through to the documentation generator
   - Update CLI help documentation

2. Enhance Type Discovery
   - Create new type discovery mode for --all
   - Implement recursive struct support
   - Add package-level type collection
   - Add generated code detection

3. Modify Error Handling
   - Create warning collection system
   - Convert error returns to warnings in --all mode
   - Implement warning reporting format
   - Ensure partial documentation is saved

4. Documentation Generation
   - Update YAML generation for package-wide types
   - Handle recursive type references
   - Preserve existing Service documentation format
   - Merge overlapping type documentation

5. Testing
   - Add unit tests for recursive types
   - Add integration tests with httpinoapi
   - Test partial documentation generation
   - Test warning system

6. Documentation
   - Update user documentation with --all flag
   - Add examples for package-wide documentation
   - Document warning messages
   - Add troubleshooting guide

Each implementation step should:
- Maintain backward compatibility
- Include unit tests
- Generate warnings instead of errors in --all mode
- Preserve partial documentation on warnings
