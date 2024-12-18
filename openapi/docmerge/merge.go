package docmerge

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/ryboe/q"
	"github.com/utrack/pontoon/docgen"
)

// mergeOptions configures the merge behavior
type mergeOptions struct {
	preserveExisting bool // preserve existing documentation in the OpenAPI Document
	strictValidation bool // error out on any potential issue during merge
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
func Merge(inDef *v3.Document, docs *docgen.YAMLDoc, opts ...Option) (*v3.Document, error) {
	options := &mergeOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if inDef == nil {
		return nil, &ErrValidation{Field: "oapi-definition", Message: "runtime document is nil"}
	}
	if docs == nil {
		return nil, &ErrValidation{Field: "docs", Message: "documentation is nil"}
	}

	// Create a copy of the runtime document to avoid modifying the original
	// TODO issues with recursions
	//mergedDoc := reprint.This(inDef).(*v3.Document)
	mergedDoc := inDef

	// Merge documentation from YAML
	if err := mergeYAMLDocs(mergedDoc, docs, options); err != nil {
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
	if doc.Components == nil || doc.Components.Schemas == nil {
		return nil
	}

	// Iterate over all schemas in OpenAPI doc
	for schemaName, schemaReadOnly := range doc.Components.Schemas.FromNewest() {

		typSchema := schemaReadOnly.Schema()

		if typSchema.Extensions == nil {
			continue
		}

		// Get Go type information
		pkgPath, ok := typSchema.Extensions.Get("x-pontoon-go-package")
		if !ok {
			continue
		}
		typeName, ok := typSchema.Extensions.Get("x-pontoon-go-type")
		if !ok {
			continue
		}

		typeDoc := findType(docs, pkgPath.Value, typeName.Value)
		if typeDoc != nil {
			if typeDoc.Comment != "" && (!opts.preserveExisting || typSchema.Description == "") {
				typSchema.Description = typeDoc.Comment
			}
		}

		// Add field documentation
		if typSchema.Properties != nil {
			for key, fieldSchemaReadOnly := range typSchema.Properties.FromNewest() {
				fieldSchema := fieldSchemaReadOnly.Schema()

				if fieldSchema.Extensions == nil {
					continue
				}

				// Get field name
				fieldName, ok := fieldSchema.Extensions.Get("x-pontoon-field-go-name")
				if !ok {
					continue
				}

				// Find field documentation
				field := findField(typeDoc, fieldName.Value)
				if field == nil {
					continue
				}

				// Add field documentation
				if field.Comment != "" && (!opts.preserveExisting || fieldSchema.Description == "") {
					fieldSchema.Description = field.Comment
				}
				typSchema.Properties.Set(key, base.CreateSchemaProxy(fieldSchema))
			}
		}
		doc.Components.Schemas.Set(schemaName, base.CreateSchemaProxy(typSchema))
	}

	q.Q("schemas after", doc.Components.Schemas)
	return nil
}

// findField finds a field in a service by its Go name
func findField(t *docgen.YAMLType, fieldName string) *docgen.YAMLField {
	for _, field := range t.Fields {
		if field.Name == fieldName {
			return &field
		}
	}
	return nil
}

// findType finds a type in docs by its Go type name
func findType(docs *docgen.YAMLDoc, pkgPath, typeName string) *docgen.YAMLType {
	for _, t := range docs.Types {
		if t.Package == pkgPath && t.Name == typeName {
			return &t
		}
	}
	return nil
}
