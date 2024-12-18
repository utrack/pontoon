# Feature: Total Description

## Overview
Total Description is a library feature that enhances OpenAPI documentation by merging runtime OpenAPI definitions (generated via `httpinmeditate`) with documentation YAML generated via `cmd/docgen`. This creates a comprehensive API documentation that combines both runtime behavior and static documentation.

## Business Value

### Developer Experience
- Single source of truth for API documentation, eliminating discrepancies between code and docs
- Improved API discoverability through comprehensive type and field descriptions
- Reduced cognitive load by having runtime behavior and documentation in one place
- Better IDE integration through accurate OpenAPI definitions

### Maintenance & Quality
- Reduces documentation maintenance overhead by automating the merge of different documentation sources
- Ensures consistency between code behavior and documentation
- Catches documentation drift early by validating docs against runtime behavior
- Enables automated documentation quality checks

### Business Operations
- Faster onboarding for new team members with complete API documentation
- Reduced support costs through better-documented APIs
- Improved API governance through consistent documentation
- Better compliance through traceable documentation-to-code mapping

### Technical Debt
- Eliminates documentation drift by keeping docs close to code
- Reduces the cost of future API changes by maintaining a single source of truth
- Enables gradual documentation improvement without breaking changes
- Facilitates API versioning and deprecation tracking

## Target Users

### Primary Users
1. API Developers
   - Need: Quick understanding of API behavior and constraints
   - Value: Immediate access to both runtime rules and documentation
   - Pain Points Solved: 
     - No more switching between code and docs
     - Clear validation rules
     - Easy discovery of API capabilities

2. API Consumers
   - Need: Clear understanding of API usage and constraints
   - Value: Complete, accurate, and up-to-date documentation
   - Pain Points Solved:
     - Trustworthy documentation
     - Clear error conditions
     - Understanding of business logic

3. Technical Writers
   - Need: Ability to enhance API docs without code changes
   - Value: Can focus on improving descriptions while maintaining accuracy
   - Pain Points Solved:
     - No risk of documentation drift
     - Clear separation of concerns
     - Easy documentation updates

### Secondary Users
1. Project Managers
   - Need: API progress and documentation quality metrics
   - Value: Clear visibility into documentation coverage

2. QA Engineers
   - Need: Understanding of expected API behavior
   - Value: Clear test requirements from merged documentation

## Requirements

### Functional Requirements
1. Merge OpenAPI runtime definitions with documentation YAML
   - Combine OpenAPI schema information from `httpinmeditate` with doc strings from `docgen`
   - Preserve all original OpenAPI metadata including:
     - Type information
     - Validation rules
     - HTTP bindings (path/query/header/cookie parameters)
   - Add documentation-specific fields from YAML:
     - Comments and descriptions
     - Service metadata
     - Type relationships
     - Field documentation
   - Support documentation versioning:
     - Track documentation updates
     - Mark deprecated features
     - Version documentation alongside code

2. Documentation Enhancement
   - Support all OpenAPI 3.1 specification fields via `github.com/pb33f/libopenapi`
   - Allow documentation YAML to extend OpenAPI fields:
     - Add descriptions to types and fields
     - Enhance parameter documentation
     - Add service-level documentation
     - Include usage examples
     - Document error conditions and handling
   - Maintain backward compatibility with existing OpenAPI tooling
   - Support all `httpin` annotation features
   - Generate documentation quality metrics:
     - Coverage statistics
     - Missing documentation reports
     - Consistency checks

### Non-Functional Requirements
1. Performance
   - Minimal runtime overhead when merging definitions
   - Efficient memory usage for large API definitions
   - Reuse existing schema references when possible
   - Documentation generation under 1 second for typical APIs

2. Maintainability
   - Clear separation between OpenAPI and documentation concerns
   - Well-documented merge strategy
   - Extensible design for future documentation sources
   - Proper error handling for inconsistencies
   - Clear upgrade path for documentation format changes

3. Usability
   - Generated documentation must be human-readable
   - Documentation must be searchable
   - Support for documentation tooling (Swagger UI, ReDoc)
   - Clear error messages for documentation issues
   - IDE integration capabilities

4. Reliability
   - Validation of documentation correctness
   - No loss of existing documentation during merges
   - Graceful handling of documentation conflicts
   - Automatic backup of documentation artifacts

## Technical Design

### Components
1. OpenAPI Generator (`openapi/httpinmeditate`)
   - Generates OpenAPI 3.1 schemas from Go types
   - Supports complex types including:
     - Nullable types (pointers)
     - Arrays and slices
     - Maps
     - Embedded structs
     - Custom types
   - Handles HTTP bindings via `httpin` tags

2. Documentation Generator (`cmd/docgen`)
   - Extracts documentation from Go source files
   - Generates YAML documentation with:
     - Service definitions
     - Type information
     - Field documentation
     - Method documentation
   - Supports documentation quality checks
   - Generates documentation coverage reports

3. Total Description Merger (to be implemented)
   - Combines OpenAPI and YAML documentation
   - Resolves type references
   - Merges field documentation
   - Maintains OpenAPI 3.1 compatibility
   - Handles documentation conflicts
   - Generates merged documentation artifacts
   - Provides documentation validation

## Dependencies
- OpenAPI definition generator (`openapi/httpinmeditate`)
- Documentation generator (`cmd/docgen`)
- OpenAPI library (`github.com/pb33f/libopenapi`)
- YAML parsing (`gopkg.in/yaml.v3`)

## Success Metrics

### Quantitative Metrics
1. Documentation Coverage
   - 100% of OpenAPI endpoints have enhanced documentation
   - 100% of public types documented
   - 100% of parameters documented
   - 100% of error conditions documented

2. Quality Metrics
   - Zero documentation inconsistencies
   - All merged documentation valid per OpenAPI 3.1
   - Documentation generation time under 1 second
   - Zero documentation conflicts

### Qualitative Metrics
1. Developer Experience
   - Positive feedback from API developers
   - Reduced questions about API usage
   - Faster API integration time

2. Documentation Quality
   - Clear and consistent documentation style
   - Comprehensive type and field descriptions
   - Useful examples for all endpoints
   - Clear error documentation

## Timeline
Created: 2024-12-18T20:59:10+01:00

## Status
- [x] Feature Description (Current)
- [ ] RFC
- [ ] Implementation
