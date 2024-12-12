# Changelog

## [Unreleased]

### Added
- Pre-compile comment extraction component for Runtime Inspection feature
  - Added `docgen` package for documentation generation tools
  - Implemented AST walker for extracting comments from Go source code
  - Added TOML generation with checksums for documentation storage
  - Created `docgen` command line tool for documentation generation

### Changed
- Migrated documentation format from TOML to YAML for better readability and maintainability
- Enhanced type documentation with support for nullable fields, arrays, and maps
- Improved handling of embedded fields and type relationships
- Added deduplication of types across services