package httpinmeditate

import (
	"fmt"
	"reflect"

	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
)

// Generator generates OpenAPI 3.1 Component Schemas from httpin-annotated structs
type Generator struct {
	// SchemaNameFunc allows customizing schema names for types
	SchemaNameFunc func(reflect.Type) string

	refs map[reflect.Type]*base.SchemaProxy
}

// NewGenerator creates a new schema generator with default options
func NewGenerator() *Generator {
	return &Generator{
		SchemaNameFunc: defaultSchemaName,
		refs:           make(map[reflect.Type]*base.SchemaProxy),
	}
}

// defaultSchemaName returns a default schema name for a type
func defaultSchemaName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.Name()
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

type options struct {
	// rootRef controls what to do if the root schema is a pointer - true if
	// it should be a reference, false if it should be a schema
	rootRef bool

	insideBody bool
}

type Option func(*options)

// WithRootAsReference controls what to do if the root schema is a pointer.
// true if it should be a reference, false if it should be a schema
func withRootAsReference(rootRef bool) Option {
	return func(o *options) {
		o.rootRef = rootRef
	}
}

func withInsideBody(yes bool) Option {
	return func(o *options) {
		o.insideBody = yes
	}
}

// GenerateSchema generates an OpenAPI 3.1 Component Schema for the given type
func (g *Generator) generateSchema(t reflect.Type, oo ...Option) (*base.SchemaProxy, error) {
	opts := options{}
	for _, o := range oo {
		o(&opts)
	}

	debugLog("generateSchema for '%v'", t.String())
	if t.Kind() == reflect.Interface {
		return nil, errors.Errorf("interface types are not supported: %v", t)
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		innerSchema, err := g.generateSchema(t.Elem())
		if err != nil {
			return nil, err
		}

		if opts.rootRef {
			innerSchema = base.CreateSchemaProxyRef(innerSchema.Schema().SchemaTypeRef)
		}

		return base.CreateSchemaProxy(&base.Schema{
			OneOf: []*base.SchemaProxy{
				base.CreateSchemaProxy(&base.Schema{Type: []string{"null"}}),
				innerSchema,
			},
		}), nil
	}

	// Return existing reference if type was already processed
	if _, ok := g.refs[t]; ok {
		debugLog("returning alias")
		return base.CreateSchemaProxyRef(fmt.Sprintf("#/components/schemas/%s", g.SchemaNameFunc(t))), nil
	}
	g.refs[t] = nil

	// Generate new schema
	schema, err := g.generateStructSchema(t, opts.insideBody)
	if err != nil {
		return nil, err
	}

	// Store reference
	g.refs[t] = schema
	return schema, nil
}

func (g *Generator) Components() *v3.Components {
	components := &v3.Components{
		Schemas: orderedmap.New[string, *base.SchemaProxy](),
	}
	for t, schema := range g.refs {
		components.Schemas.Set(g.SchemaNameFunc(t), schema)
	}
	return components
}

// generateStructSchema generates a schema for a struct type
func (g *Generator) generateStructSchema(t reflect.Type, isJSON bool) (*base.SchemaProxy, error) {
	if t.Kind() != reflect.Struct {
		return nil, errors.Errorf("type must be a struct, got %v", t.Kind())
	}

	debugLog("genStructSchema for '%v'", t.String())

	// Handle embedded fields first - we'll merge them via allOf
	var embeddedSchemas []*base.SchemaProxy
	var regularFields []*fieldInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous {
			embeddedType := field.Type

			schema, err := g.generateSchema(embeddedType, withRootAsReference(true))
			if err != nil {
				return nil, errors.Wrapf(err, "when generating schema for embedded field '%v' of type '%v'", field.Name, field.Type)
			}

			var embeddedSchema *base.SchemaProxy
			switch {
			case schema.IsReference():
				embeddedSchema = base.CreateSchemaProxyRef(schema.GetReference())
			case schema.Schema() != nil && schema.Schema().SchemaTypeRef != "":
				embeddedSchema = base.CreateSchemaProxyRef(schema.Schema().SchemaTypeRef)
			default:
				embeddedSchema = schema
			}
			embeddedSchemas = append(embeddedSchemas, embeddedSchema)
			continue
		}

		info, err := parseField(field, isJSON)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing field %s", field.Name)
		}
		debugLog("--> info for the field: %+v", info)
		if info != nil {
			regularFields = append(regularFields, info)
		}
	}

	// Create schema for regular fields
	propMap := orderedmap.New[string, *base.SchemaProxy]()
	for _, field := range regularFields {
		fieldSchema, err := g.generateFieldSchema(field)
		debugLog("fieldSchema '%v': %v", field.Name, fieldSchema)
		if err != nil {
			return nil, errors.Wrapf(err, "error generating schema for field %s", field.Name)
		}
		propMap.Set(field.Name, fieldSchema)
	}

	regularSchema := base.CreateSchemaProxy(&base.Schema{
		SchemaTypeRef: fmt.Sprintf("#/components/schemas/%s", g.SchemaNameFunc(t)),
		Type:          []string{"object"},
		Properties:    propMap,
	})

	// If we have embedded fields, use allOf
	if len(embeddedSchemas) > 0 {
		allOf := append([]*base.SchemaProxy{}, embeddedSchemas...)
		allOf = append(allOf, regularSchema)
		return base.CreateSchemaProxy(&base.Schema{
			SchemaTypeRef: fmt.Sprintf("#/components/schemas/%s", g.SchemaNameFunc(t)),
			Type:          []string{"object"},
			AllOf:         allOf,
		}), nil
	}

	return regularSchema, nil
}

// generateFieldSchema generates a schema for a field
func (g *Generator) generateFieldSchema(field *fieldInfo) (*base.SchemaProxy, error) {
	fieldType := field.Type
	debugLog("proc field '%v' %v", field.Name, field.Type.String())

	// Create extensions map from httpin tags
	extensions := orderedmap.New[string, *yaml.Node]()
	if field.In != "" {
		extensions.Set("x-httpin-in", &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: field.In,
		})
	}
	for tagType, tagValue := range field.Tags {
		extensions.Set(fmt.Sprintf("x-httpin-%s", tagType), &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: tagValue,
		})
	}

	// Handle pointer types first
	if fieldType.Kind() == reflect.Ptr {
		debugLog("field is ptr: '%v' %v", field.Name, field.Type.String())
		innerSchema, err := g.generateFieldSchema(&fieldInfo{
			Name: field.Name,
			In:   field.In,
			Type: fieldType.Elem(),
		})
		if err != nil {
			return nil, err
		}

		if innerSchema.IsReference() {
			//g.refs[fieldType.Elem()] = innerSchema
			//innerSchema = base.CreateSchemaProxyRef(innerSchema.GetReference())
		}

		schema := &base.Schema{
			OneOf: []*base.SchemaProxy{
				base.CreateSchemaProxy(&base.Schema{Type: []string{"null"}}),
				innerSchema,
			},
		}
		if extensions.Len() > 0 {
			schema.Extensions = extensions
		}
		return base.CreateSchemaProxy(schema), nil
	}

	// Handle primitive types
	primSchema, err := primitiveTypeToSchema(fieldType)
	if err != nil {
		return nil, errors.Wrap(err, "error converting primitive type")
	}
	if primSchema != nil {
		debugLog("-> primitive field")
		if extensions.Len() > 0 {
			primSchema.Extensions = extensions
		}
		return base.CreateSchemaProxy(primSchema), nil
	}

	// Handle complex types
	switch fieldType.Kind() {
	case reflect.Struct:
		debugLog(" -> struct field")

		structSchema, err := g.generateSchema(fieldType, withInsideBody(field.In == "body"))
		if err != nil {
			return nil, err
		}

		if structSchema.IsReference() {
			debugLog("  -> is reference")
			schema := &base.Schema{
				SchemaTypeRef: structSchema.GetReference(),
			}
			if extensions.Len() > 0 {
				schema.Extensions = extensions
			}
			return base.CreateSchemaProxy(schema), nil
		}

		if structSchema.Schema() != nil && structSchema.Schema().SchemaTypeRef != "" {
			debugLog("  -> ref != nil")
			schema := &base.Schema{
				SchemaTypeRef: structSchema.Schema().SchemaTypeRef,
			}
			if extensions.Len() > 0 {
				schema.Extensions = extensions
			}
			return base.CreateSchemaProxy(schema), nil
		}

		if extensions.Len() > 0 && structSchema.Schema() != nil {
			structSchema.Schema().Extensions = extensions
		}
		return structSchema, nil

	case reflect.Slice, reflect.Array:
		itemSchema, err := g.generateFieldSchema(&fieldInfo{
			Name: field.Name,
			Type: fieldType.Elem(),
		})
		if err != nil {
			return nil, err
		}

		if itemSchema.IsReference() {
			itemSchema = base.CreateSchemaProxyRef(itemSchema.GetReference())
		}
		schema := &base.Schema{
			Type:  []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: itemSchema, N: 0},
		}
		if extensions.Len() > 0 {
			schema.Extensions = extensions
		}
		return base.CreateSchemaProxy(schema), nil

	case reflect.Map:
		if fieldType.Key().Kind() != reflect.String {
			return nil, errors.Errorf("map key must be string, got %v", fieldType.Key().Kind())
		}
		valueSchema, err := g.generateFieldSchema(&fieldInfo{
			Name: field.Name,
			Type: fieldType.Elem(),
		})
		if err != nil {
			return nil, err
		}
		if valueSchema.IsReference() {
			valueSchema = base.CreateSchemaProxyRef(valueSchema.GetReference())
		}
		schema := &base.Schema{
			Type:                 []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{A: valueSchema, N: 0},
		}
		if extensions.Len() > 0 {
			schema.Extensions = extensions
		}
		return base.CreateSchemaProxy(schema), nil

	case reflect.Interface:
		return nil, errors.New("interface types are not supported")

	default:
		return nil, errors.Errorf("unsupported type: %v", fieldType.Kind())
	}
}
