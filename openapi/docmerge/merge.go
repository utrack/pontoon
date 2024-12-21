package docmerge

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pkg/errors"
	"github.com/utrack/pontoon/docgen"
	"github.com/utrack/pontoon/docgen/docregistry"
	ext "github.com/utrack/pontoon/openapi/pontoonext"
)

// mergeOptions configures the merge behavior
type mergeOptions struct {
	preserveExisting bool // preserve existing documentation in the OpenAPI Document
	strictValidation bool // error out on any potential issue during merge
	docFile          *docgen.YAMLDoc
}

// Option is a functional option for merge configuration
type Option func(*mergeOptions)

// WithPreserveExisting configures whether to keep existing documentation in the OpenAPI Document.
// If true, existing documentation will not be overwritten by the YAML documentation.
func WithPreserveExisting(preserve bool) Option {
	return func(o *mergeOptions) {
		o.preserveExisting = preserve
	}
}

// WithStrictValidation enables strict validation during merge.
// If true, any potential issue (missing docs, type mismatches etc) will result in an error.
func WithStrictValidation(strict bool) Option {
	return func(o *mergeOptions) {
		o.strictValidation = strict
	}
}

// WithDocFile uses a given docfile instead of the global registry.
// Useful for testing internals, not much else.
func WithDocFile(f *docgen.YAMLDoc) Option {
	return func(o *mergeOptions) {
		o.docFile = f
	}
}

// ErrValidation represents a validation error during merge
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	return "validation error in " + e.Field + ": " + e.Message
}

// ErrMergeConflict represents a conflict between runtime and documentation
type ErrMergeConflict struct {
	Path    string
	Runtime interface{}
	Doc     interface{}
}

func (e *ErrMergeConflict) Error() string {
	return "merge conflict at " + e.Path
}

// Merge merges the documentation into the OpenAPI schema
func Merge(inDef *v3.Document, opts ...Option) (*v3.Document, error) {
	options := &mergeOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if inDef == nil {
		return nil, &ErrValidation{Field: "oapi-definition", Message: "runtime document is nil"}
	}
	if options.docFile == nil {
		options.docFile = docregistry.GlobalFile()
	}

	// Create a copy of the runtime document to avoid modifying the original
	// TODO issues with recursions
	//mergedDoc := reprint.This(inDef).(*v3.Document)
	mergedDoc := inDef

	// Merge documentation from YAML
	if err := mergeYAMLDocs(mergedDoc, options.docFile, options); err != nil {
		return nil, err
	}

	// Create and configure the pipeline
	pipeline := NewPipeline(
		NewCommentProcessor(),
		// Add more stages here as needed
	)

	// Process the document through the pipeline
	if err := pipeline.Process(mergedDoc); err != nil {
		return nil, err
	}

	return mergedDoc, nil
}

// mergeYAMLDocs merges documentation from YAML into OpenAPI document
func mergeYAMLDocs(doc *v3.Document, docs *docgen.YAMLDoc, opts *mergeOptions) error {

	if doc.Components != nil && doc.Components.Schemas != nil {
		err := mergeDocComponents(doc.Components.Schemas, docs, opts)
		if err != nil {
			return errors.Wrap(err, "when merging Components.Schemas")
		}
	}

	// TODO merge services and handlers

	return nil
}

// mergeDocComponents merges documentation from YAML into OpenAPI `Components.Schemas`.
func mergeDocComponents(schemas *orderedmap.Map[string, *base.SchemaProxy], docs *docgen.YAMLDoc, opts *mergeOptions) error {
	for schemaName, schemaReadOnly := range schemas.FromNewest() {

		schema := schemaReadOnly.Schema()
		err := mergeSchemaDoc(schema, docs, opts)
		if err != nil {
			return errors.Wrapf(err, "when merging schema '%v'", schemaName)
		}

		schemas.Set(schemaName, base.CreateSchemaProxy(schema))
	}
	return nil
}

// mergeCompositeSchemas processes all schemas in a composite type (allOf/anyOf/oneOf)
func mergeCompositeSchemas(schemas []*base.SchemaProxy, docs *docgen.YAMLDoc, opts *mergeOptions) error {
	for i, partReadOnly := range schemas {
		if partReadOnly == nil {
			continue
		}
		schema := partReadOnly.Schema()
		if err := mergeSchemaDoc(schema, docs, opts); err != nil {
			return fmt.Errorf("failed to merge schema at index %d: %w", i, err)
		}
		schemas[i] = base.CreateSchemaProxy(schema)
	}
	return nil
}

func mergeSchemaDoc(schema *base.Schema, docs *docgen.YAMLDoc, opts *mergeOptions) error {
	if schema.Extensions == nil {
		return nil
	}

	// Get Go type information
	goTypeInfo, ok := ext.GetGoTypeInfo(schema.Extensions)
	if !ok {
		return nil
	}

	if err := mergeCompositeSchemas(schema.AllOf, docs, opts); err != nil {
		return errors.Wrapf(err, "when merging AllOf for Go type '%v'", goTypeInfo.String())
	}

	if err := mergeCompositeSchemas(schema.AnyOf, docs, opts); err != nil {
		return errors.Wrapf(err, "when merging AnyOf for Go type '%v'", goTypeInfo.String())
	}

	if err := mergeCompositeSchemas(schema.OneOf, docs, opts); err != nil {
		return errors.Wrapf(err, "when merging OneOf for Go type '%v'", goTypeInfo.String())
	}

	typeDoc := findType(docs, goTypeInfo.PackagePath, goTypeInfo.TypeName)
	if typeDoc != nil {
		if typeDoc.Comment != "" && (!opts.preserveExisting || schema.Description == "") {
			schema.Description = typeDoc.Comment
		}
	}

	// Add field documentation
	if schema.Properties == nil || schema.Properties.Len() == 0 {
		return nil
	}
	for key, fieldSchemaReadOnly := range schema.Properties.FromNewest() {
		fieldSchema := fieldSchemaReadOnly.Schema()

		if fieldSchema.Extensions == nil {
			continue
		}

		// Get field name
		fieldName, ok := ext.GetGoFieldName(fieldSchema.Extensions)
		if !ok {
			continue
		}

		// Find field documentation
		field := findField(typeDoc, fieldName)
		if field == nil {
			continue
		}

		// Add field documentation
		if field.Comment != "" && (!opts.preserveExisting || fieldSchema.Description == "") {
			fieldSchema.Description = field.Comment
		}
		schema.Properties.Set(key, base.CreateSchemaProxy(fieldSchema))
	}

	return nil
}

// findField finds a field in a service by its Go name
func findField(t *docgen.YAMLType, fieldName string) *docgen.YAMLField {
	if t == nil {
		return nil
	}
	for _, field := range t.Fields {
		if field.Name == fieldName {
			return &field
		}
	}
	return nil
}

// findType finds a type in docs by its Go type name
func findType(docs *docgen.YAMLDoc, pkgPath, typeName string) *docgen.YAMLType {
	if docs == nil {
		return nil
	}
	v,ok := docs.Types[pkgPath+"."+typeName]
	if !ok {
		return nil
	}
	return &v
}
