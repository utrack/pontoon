# Runtime Inspection

## Overview
Currently, Pontoon's documentation is generated at pre-compile time, which has several limitations:
- Highly dependent on the exact setup
- Inflexible and clunky
- Lacks runtime information

This feature aims to improve documentation generation by combining the strengths of compile-time introspection with runtime flexibility.

## Components

### 1. Compile-time Documentation Generation
- Extract parsed comments from:
  - Handler implementations (structs implementing `sdesc.Service`)
  - Input/output structs and their fields
  - Related type definitions
- Extract service-level documentation and options
- Document HTTP routing patterns and methods
- Capture error types and conditions
- Generate static documentation that captures design-time information
- Store documentation as TOML in Go source files as constants
  - Enables git-based change tracking
  - Makes regeneration detection easier
- Preserve documentation hierarchy and relationships between types

### 2. Runtime Model Generation
- Generate handler models during runtime
- Capture input/output descriptions with full type information
- Include actual runtime types and configurations
- Provide real-time validation rules and constraints
- Maintain relationship between request/response structs and their fields
- Generate documentation for service configurations and options
- Include middleware effects and transformations
- Document error scenarios and validation rules

### 3. Documentation Merger
- Combine compile-time documentation with runtime models
- Preserve both static documentation and dynamic information
- Create a unified view that leverages both sources
- Ensure all comments (handler, struct, and field-level) are preserved and properly linked
- Error handling:
  - Return errors if compile-time and runtime data conflicts are detected
  - Conflicts indicate stale generated code that needs regeneration
  - Fail-fast approach to prevent documentation inconsistencies

## Benefits
- More reliable documentation that reflects actual runtime behavior
- Better type information and validation rules
- Complete documentation coverage from high-level handlers down to individual fields
- Reduced dependency on build-time configuration
- Improved developer experience with accurate API documentation

## Implementation Steps

1. Compile-time Parser
   - Create a parser for all Go comments (handlers, structs, fields)
   - Extract service-level documentation
   - Parse error types and conditions
   - Document HTTP routing configuration
   - Generate Go structs to hold documentation
   - Generate TOML-formatted documentation constants
   - Build a documentation graph to maintain relationships
   - Add hash/checksum for regeneration detection

2. Runtime Model Generator
   - Create runtime type introspection system
   - Detect and parse `sdesc.Service` implementations
   - Generate handler models with full type information
   - Generate service configuration documentation
   - Document middleware effects
   - Capture error scenarios
   - Implement validation rule extraction
   - Preserve type relationships and hierarchies

3. Documentation Merger
   - Design merge algorithm
   - Implement strict validation between compile-time TOML and runtime data
   - Return descriptive errors for documentation conflicts
   - Generate final documentation format
   - Ensure proper linking between all documentation levels

## Technical Considerations
- Need to ensure minimal runtime overhead
- Consider caching strategies for runtime model generation
- Design a flexible merge system that preserves important information from both sources
- Plan for backward compatibility with existing documentation
- Handle nested and referenced types properly
- Determine documentation generation timing (startup vs on-demand)
- Monitor and optimize memory usage for documentation storage
- Consider startup time impact
- TOML parsing performance impact at startup
- Git-friendly TOML formatting for better diffing

## Future Possibilities
- Interactive API documentation with live examples
- Runtime validation based on merged documentation
- Integration with API testing tools
- Smart documentation diffing to highlight changes in types and their documentation

## Development Notes

### Key Decisions
1. **Storage Format**: TOML inside Go constants
   - Enables git-based change tracking
   - Makes it easy to detect when regeneration is needed
   - Provides good readability for code review

2. **Handler Detection**:
   - Handlers are structs implementing `sdesc.Service` interface
   - No additional annotations or markers needed

3. **Documentation Merging Strategy**:
   - Strict validation between compile-time and runtime data
   - Fail-fast on any conflicts
   - Conflicts indicate need for documentation regeneration

### Implementation Priorities
1. Compile-time parser implementation
2. Runtime model generation
3. Documentation merger with conflict detection

### Future Considerations
- This is a large feature that may be split into smaller parts before the RFC phase
- Each component (parser, runtime generation, merger) could potentially be its own feature
