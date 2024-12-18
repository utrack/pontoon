package httpinmeditate

import (
	"encoding"
	"fmt"
	"reflect"
	"testing"
	"time"

	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	debugEnabled = true
}

type Address struct {
	Street string `httpin:"form=street"`
	City   string `httpin:"form=city"`
}

type User struct {
	Address                     // embedded struct
	ID        int64             `httpin:"path=id"`
	Name      *string           `httpin:"form=name"`
	Age       int               `httpin:"query=age"`
	Tags      []string          `httpin:"form=tags"`
	Metadata  map[string]string `httpin:"json=metadata"`
	CreatedAt time.Time         `httpin:"header=created-at"`
	UpdatedAt *time.Time        `httpin:"header=updated-at"`
	private   string            // should be ignored
}

type CustomString string
type CustomInt int
type CustomTime time.Time

type MarshalableType struct {
	value string
}

func (m MarshalableType) MarshalText() ([]byte, error) {
	return []byte(m.value), nil
}

var _ encoding.TextMarshaler = (*MarshalableType)(nil)

type Node struct {
	Value    string `httpin:"form=value"`
	Parent   *Node  `httpin:"form=parent"`
	Children []Node `httpin:"form=children"`
}

type BaseA struct {
	FieldA string `httpin:"form=field_a"`
}

type BaseB struct {
	FieldB string `httpin:"form=field_b"`
}

type MultiEmbed struct {
	BaseA
	*BaseB
	FieldC string `httpin:"form=field_c"`
}

type ComplexTypes_SlicePtr struct {
	NestedPtrSlice []*BaseA `httpin:"form=nested_ptr_slice"`
}

type ComplexTypes_StructMap struct {
	StructMap map[string]BaseA `httpin:"form=struct_map"`
}

type ComplexTypes struct {
	String string  `httpin:"form=string"`
	Int    int     `httpin:"form=int"`
	Float  float64 `httpin:"form=float"`
	Bool   bool    `httpin:"form=bool"`
	Bytes  []byte  `httpin:"form=bytes"`

	CustomStr  CustomString    `httpin:"form=custom_str"`
	CustomInt  CustomInt       `httpin:"form=custom_int"`
	CustomTime CustomTime      `httpin:"form=custom_time"`
	Marshaler  MarshalableType `httpin:"form=marshaler"`

	StringPtr *string  `httpin:"form=string_ptr"`
	IntPtr    *int     `httpin:"form=int_ptr"`
	FloatPtr  *float64 `httpin:"form=float_ptr"`
	BoolPtr   *bool    `httpin:"form=bool_ptr"`

	StringArray  [3]string `httpin:"form=string_array"`
	IntSlice     []int     `httpin:"form=int_slice"`
	Float64Slice []float64 `httpin:"form=float_slice"`

	StringMap map[string]string `httpin:"form=string_map"`
	IntMap    map[string]int    `httpin:"form=int_map"`
	StructMap map[string]BaseA  `httpin:"form=struct_map"`

	Nested         BaseA    `httpin:"form=nested"`
	NestedPtr      *BaseA   `httpin:"form=nested_ptr"`
	NestedSlice    []BaseA  `httpin:"form=nested_slice"`
	NestedPtrSlice []*BaseA `httpin:"form=nested_ptr_slice"`

	Time         time.Time    `httpin:"form=time"`
	TimePtr      *time.Time   `httpin:"form=time_ptr"`
	TimeSlice    []time.Time  `httpin:"form=time_slice"`
	TimePtrSlice []*time.Time `httpin:"form=time_ptr_slice"`

	EmptyStruct struct{} `httpin:"form=empty_struct"`
	// AnonStruct  struct {
	// 	Field string `httpin:"form=field"`
	// } `httpin:"form=anon_struct"`
}

func TestGenerator_GenerateSchema(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(User{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	// Convert schema to map for easier testing
	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)

	// Check embedded struct
	require.NotNil(t, doc.AllOf)
	require.Len(t, doc.AllOf, 2) // Address + regular fields

	// Check Address schema
	addrReference := doc.AllOf[0].GetReference()
	assert.Equal(t, "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.Address", addrReference)

	// Check regular fields
	regularSchema := doc.AllOf[1].Schema()
	assert.Equal(t, []string{"object"}, regularSchema.Type)
	require.NotNil(t, regularSchema.Properties)
	props := regularSchema.Properties

	// Check ID field
	idSchema := props.GetOrZero("id").Schema()
	fmt.Println(idSchema)
	assert.Equal(t, []string{"integer"}, idSchema.Type)

	// Check Name field (nullable)
	nameSchema := props.GetOrZero("name").Schema()
	require.NotNil(t, nameSchema.OneOf)
	assert.Len(t, nameSchema.OneOf, 2)
	assert.Equal(t, []string{"null"}, nameSchema.OneOf[0].Schema().Type)
	assert.Equal(t, []string{"string"}, nameSchema.OneOf[1].Schema().Type)

	// Check Tags field (array)
	tagsSchema := props.GetOrZero("tags").Schema()
	assert.Equal(t, []string{"array"}, tagsSchema.Type)
	assert.Equal(t, []string{"string"}, tagsSchema.Items.A.Schema().Type)

	// Check Metadata field (map)
	metaSchema := props.GetOrZero("metadata").Schema()
	assert.Equal(t, []string{"object"}, metaSchema.Type)
	require.NotNil(t, metaSchema.AdditionalProperties)

	// Check CreatedAt field (time.Time)
	createdSchema := props.GetOrZero("created-at").Schema()
	assert.Equal(t, []string{"string"}, createdSchema.Type)
	assert.Equal(t, "date-time", createdSchema.Format)

	// Check UpdatedAt field (nullable time.Time)
	updatedSchema := props.GetOrZero("updated-at").Schema()
	require.NotNil(t, updatedSchema.OneOf)
	assert.Len(t, updatedSchema.OneOf, 2)
	assert.Equal(t, []string{"null"}, updatedSchema.OneOf[0].Schema().Type)
	timeSchema := updatedSchema.OneOf[1].Schema()
	assert.Equal(t, []string{"string"}, timeSchema.Type)
	assert.Equal(t, "date-time", timeSchema.Format)
}

func TestGenerator_GenerateSchema_Errors(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		name     string
		typ      reflect.Type
		wantErr  bool
		errMatch string
	}{
		{
			name:     "interface type",
			typ:      reflect.TypeOf((*interface{})(nil)).Elem(),
			wantErr:  true,
			errMatch: "interface types are not supported",
		},
		{
			name: "struct with interface field",
			typ: reflect.TypeOf(struct {
				Field interface{} `httpin:"form=field"`
			}{}),
			wantErr:  true,
			errMatch: "type interface {} not supported",
		},
		{
			name: "map with non-string key",
			typ: reflect.TypeOf(struct {
				Field map[int]string `httpin:"form=field"`
			}{}),
			wantErr:  true,
			errMatch: "map key must be string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := g.generateSchema(tt.typ)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerator_GenerateSchema_ComplexTypes_SlicePtr(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(ComplexTypes_SlicePtr{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)

	props := doc.Properties
	require.NotNil(t, props)

	// Test basic types
	assertArrayNullableRefType(t, props, "nested_ptr_slice", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")
}

func TestGenerator_GenerateSchema_ComplexTypes_StructMap(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(ComplexTypes_StructMap{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)

	props := doc.Properties
	require.NotNil(t, props)

	assertMapRefType(t, props, "struct_map", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")
}

func TestGenerator_GenerateSchema_ComplexTypes(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(ComplexTypes{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)

	props := doc.Properties
	require.NotNil(t, props)

	// Test basic types
	assertSchemaType(t, props, "string", "string")
	assertSchemaType(t, props, "int", "integer")
	assertSchemaType(t, props, "float", "number")
	assertSchemaType(t, props, "bool", "boolean")
	assertSchemaType(t, props, "bytes", "string", "binary")

	// Test custom types
	assertSchemaType(t, props, "custom_str", "string")
	assertSchemaType(t, props, "custom_int", "integer")
	assertSchemaType(t, props, "custom_time", "string", "date-time")
	assertSchemaType(t, props, "marshaler", "string")

	// Test pointers (nullable types)
	assertNullableType(t, props, "string_ptr", "string")
	assertNullableType(t, props, "int_ptr", "integer")
	assertNullableType(t, props, "float_ptr", "number")
	assertNullableType(t, props, "bool_ptr", "boolean")

	// Test arrays and slices
	assertArrayType(t, props, "string_array", "string")
	assertArrayType(t, props, "int_slice", "integer")
	assertArrayType(t, props, "float_slice", "number")

	// Test maps
	assertMapValueType(t, props, "string_map", "string")
	assertMapValueType(t, props, "int_map", "integer")
	assertMapRefType(t, props, "struct_map", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")

	// Test nested structs
	assertRefType(t, props, "nested", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")
	assertNullableRefType(t, props, "nested_ptr", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")
	assertArrayRefType(t, props, "nested_slice", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")
	assertArrayNullableRefType(t, props, "nested_ptr_slice", "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA")

	// Test time types
	assertSchemaType(t, props, "time", "string", "date-time")
	assertNullableType(t, props, "time_ptr", "string", "date-time")
	assertArrayType(t, props, "time_slice", "string", "date-time")
	assertArrayNullableType(t, props, "time_ptr_slice", "string", "date-time")
}

func TestGenerator_GeneratableSpec(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(ComplexTypes{}))
	require.NoError(t, err)
	require.NotNil(t, schema)
	comps := orderedmap.New[string, *base.SchemaProxy]()
	comps.Set("foo", schema)

	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title: "Useless API",
			Contact: &base.Contact{
				Name:  "quobix",
				Email: "buckaroo@pb33f.io",
			},
		},
		Components: &v3.Components{
			Schemas: comps,
		},
	}
	rend, err := doc.Render()
	require.Nil(t, err)
	fmt.Print(string(rend))

}

func TestGenerator_GenerateSchema_Recursive(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(Node{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)

	props := doc.Properties
	require.NotNil(t, props)

	// Check Value field
	assertSchemaType(t, props, "value", "string")

	// Check Parent field (recursive pointer)
	parentSchema := props.GetOrZero("parent").Schema()
	require.NotNil(t, parentSchema.OneOf)
	require.Len(t, parentSchema.OneOf, 2)
	assert.Equal(t, []string{"null"}, parentSchema.OneOf[0].Schema().Type)
	assert.Equal(t, "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.Node", parentSchema.OneOf[1].Schema().SchemaTypeRef)

	// Check Children field (recursive slice)
	childrenSchema := props.GetOrZero("children").Schema()
	assert.Equal(t, []string{"array"}, childrenSchema.Type)
	assert.Equal(t, "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.Node", childrenSchema.Items.A.Schema().SchemaTypeRef)
}

func TestGenerator_GenerateSchema_MultipleEmbedded(t *testing.T) {
	g := NewGenerator()

	schema, err := g.generateSchema(reflect.TypeOf(MultiEmbed{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)
	require.NotNil(t, doc.AllOf)
	require.Len(t, doc.AllOf, 3) // BaseA + BaseB + own fields

	// Check BaseA fields
	baseARef := doc.AllOf[0].GetReference()
	assert.Equal(t, "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseA", baseARef)

	// Check BaseB fields (pointer)
	baseBPtrSchema := doc.AllOf[1].Schema()
	require.Len(t, baseBPtrSchema.OneOf, 2)
	assert.Equal(t, []string{"null"}, baseBPtrSchema.OneOf[0].Schema().Type)
	assert.Equal(t, "#/components/schemas/github.com/utrack/pontoon/openapi/httpinmeditate.BaseB", baseBPtrSchema.OneOf[1].GetReference())

	// Check own fields
	ownSchema := doc.AllOf[2].Schema()
	assert.Equal(t, []string{"object"}, ownSchema.Type)
	assertSchemaType(t, ownSchema.Properties, "field_c", "string")
}

func TestGenerator_GenerateSchema_HttpinTags(t *testing.T) {
	type TestStruct struct {
		FormField   string            `httpin:"form=form_field"`
		QueryField  *string           `httpin:"query=query_field"`
		HeaderField []int             `httpin:"header=header-field"`
		CookieField bool              `httpin:"cookie=cookie_field"`
		PathField   int64             `httpin:"path=path_field"`
		JSONField   map[string]string `httpin:"json=json_field"`
	}

	g := NewGenerator()
	schema, err := g.generateSchema(reflect.TypeOf(TestStruct{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	props := schema.Schema().Properties
	require.NotNil(t, props)

	// Check form field
	formField, ok := props.Get("form_field")
	require.True(t, ok)
	require.Equal(t, "form_field", formField.Schema().Extensions.GetOrZero("x-httpin-form").Value)

	// Check query field with pointer
	queryField, ok := props.Get("query_field")
	require.True(t, ok)
	require.Equal(t, "query_field", queryField.Schema().Extensions.GetOrZero("x-httpin-query").Value)

	// Check header field with array
	headerField, ok := props.Get("header-field")
	require.True(t, ok)
	require.Equal(t, "header-field", headerField.Schema().Extensions.GetOrZero("x-httpin-header").Value)

	// Check cookie field
	cookieField, ok := props.Get("cookie_field")
	require.True(t, ok)
	require.Equal(t, "cookie_field", cookieField.Schema().Extensions.GetOrZero("x-httpin-cookie").Value)

	// Check path field
	pathField, ok := props.Get("path_field")
	require.True(t, ok)
	require.Equal(t, "path_field", pathField.Schema().Extensions.GetOrZero("x-httpin-path").Value)

	// Check json field with map
	jsonField, ok := props.Get("json_field")
	require.True(t, ok)
	require.Equal(t, "json_field", jsonField.Schema().Extensions.GetOrZero("x-httpin-json").Value)
}

type StructWithJSON struct {
	FieldJSON StructWithJSONField `httpin:"body=json"`
}

type StructWithJSONField struct {
	Field  string `json:"field"`
	Field2 string
	Field3 string `json:"json_field3"`
}

func TestGenerator_GenerateSchema_JSONTags(t *testing.T) {
	g := NewGenerator()
	schema, err := g.generateSchema(reflect.TypeOf(StructWithJSON{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	props := schema.Schema().Properties
	require.NotNil(t, props)

	jsonField, ok := props.Get("FieldJSON")
	require.True(t, ok)
	require.Equal(t, "json", jsonField.Schema().Extensions.GetOrZero("x-httpin-body").Value)

	ref := "github.com/utrack/pontoon/openapi/httpinmeditate.StructWithJSONField"
	refString := "#/components/schemas/" + ref

	require.Equal(t, refString, jsonField.Schema().SchemaTypeRef)

	dict := g.Components()
	nestedStructDesc, ok := dict.Schemas.Get(ref)
	require.True(t, ok)

	require.Equal(t, []string{"object"}, nestedStructDesc.Schema().Type)

	sch := nestedStructDesc.Schema()
	require.Equal(t, sch.Properties.Len(), 3)
}

type ParamModel struct {
	ID      int    `httpin:"path=id"`
	Query   string `httpin:"query=q"`
	Header  string `httpin:"header=x-custom"`
	Cookie  string `httpin:"cookie=session"`
	User    User   `httpin:"body=json"`
	Nested  ParamModelNested
	Ignored string
}

type ParamModelNested struct {
	Field1 string `httpin:"path=field1"`
	Field2 string `httpin:"query=field2"`
}

func TestGenerator_GenerateOperationModel(t *testing.T) {
	g := NewGenerator()

	schema, params, err := g.GenerateModel(reflect.TypeOf(ParamModel{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title: "Useless API",
			Contact: &base.Contact{
				Name:  "quobix",
				Email: "buckaroo@pb33f.io",
			},
		},
		Components: g.Components(),
	}
	rend, err := doc.Render()
	require.Nil(t, err)
	fmt.Print(string(rend))
	// Check parameters
	require.Len(t, params, 6, "should include all parameters including nested structs")

	// Verify each parameter
	for _, p := range params {

		switch p.Name {
		case "id":
			assert.Equal(t, "path", p.In)
			assert.Equal(t, []string{"integer"}, p.Schema.Schema().Type)
			annot, ok := p.Extensions.Get("x-pontoon-field-go-name")
			assert.True(t, ok)
			assert.NotNil(t, annot)
			annot, ok = p.Extensions.Get("x-pontoon-go-package")
			require.True(t, ok)
			require.NotNil(t, annot)
			annot, ok = p.Extensions.Get("x-pontoon-go-type")
			require.True(t, ok)
			require.NotNil(t, annot)
		case "q":
			assert.Equal(t, "query", p.In)
			assert.Equal(t, []string{"string"}, p.Schema.Schema().Type)
		case "x-custom":
			assert.Equal(t, "header", p.In)
			assert.Equal(t, []string{"string"}, p.Schema.Schema().Type)
		case "session":
			assert.Equal(t, "cookie", p.In)
			assert.Equal(t, []string{"string"}, p.Schema.Schema().Type)
		case "field1":
			assert.Equal(t, "path", p.In)
			assert.Equal(t, []string{"string"}, p.Schema.Schema().Type)
		case "field2":
			assert.Equal(t, "query", p.In)
			assert.Equal(t, []string{"string"}, p.Schema.Schema().Type)
		default:
			t.Errorf("unexpected parameter: %s", p.Name)
		}
	}

}

func assertSchemaType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, typ string, format ...string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{typ}, schema.Type)
	if len(format) > 0 {
		assert.Equal(t, format[0], schema.Format)
	}
}

func assertNullableType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, typ string, format ...string) {
	schema := props.GetOrZero(field).Schema()
	require.NotNil(t, schema.OneOf)
	assert.Equal(t, []string{"null"}, schema.OneOf[0].Schema().Type)
	assert.Equal(t, []string{typ}, schema.OneOf[1].Schema().Type)
	if len(format) > 0 {
		assert.Equal(t, format[0], schema.OneOf[1].Schema().Format)
	}
}

func assertArrayType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, itemType string, format ...string) {
	fmt.Printf("assertArrayType: checking field %q (expected item type: %q, format: %v)\n", field, itemType, format)
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"array"}, schema.Type)
	itemSchema := schema.Items.A.Schema()
	assert.Equal(t, []string{itemType}, itemSchema.Type)
	if len(format) > 0 {
		assert.Equal(t, format[0], itemSchema.Format)
	}
}

func assertMapValueType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, valueType string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"object"}, schema.Type)
	valueSchema := schema.AdditionalProperties.A.Schema()
	assert.Equal(t, []string{valueType}, valueSchema.Type)
}

func assertMapRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"object"}, schema.Type)
	require.NotNil(t, schema.AdditionalProperties)
	require.NotNil(t, schema.AdditionalProperties.A)
	require.Equal(t, ref, schema.AdditionalProperties.A.Schema().SchemaTypeRef)
}

func assertRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	fmt.Println(field)
	prop := props.GetOrZero(field)
	require.NotNil(t, prop)
	refSchema := props.GetOrZero(field).Schema().SchemaTypeRef
	assert.Equal(t, ref, refSchema)
}

func assertNullableRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	schema := props.GetOrZero(field).Schema()
	require.NotNil(t, schema.OneOf)
	assert.Equal(t, []string{"null"}, schema.OneOf[0].Schema().Type)
	assert.Equal(t, ref, schema.OneOf[1].Schema().SchemaTypeRef)
}

func assertArrayRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"array"}, schema.Type)
	assert.Equal(t, ref, schema.Items.A.Schema().SchemaTypeRef)
}

func assertArrayNullableRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"array"}, schema.Type)
	itemSchema := schema.Items.A
	require.NotNil(t, itemSchema.Schema().OneOf)
	assert.Equal(t, []string{"null"}, itemSchema.Schema().OneOf[0].Schema().Type)
	elem := itemSchema.Schema().OneOf[1]
	if elem.IsReference() {
		assert.Equal(t, ref, elem.GetReference())
	} else {
		assert.Equal(t, ref, elem.Schema().SchemaTypeRef)
	}
}

func assertArrayNullableType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, typ string, format ...string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"array"}, schema.Type)
	itemSchema := schema.Items.A
	require.NotNil(t, itemSchema.Schema().OneOf)
	assert.Equal(t, []string{"null"}, itemSchema.Schema().OneOf[0].Schema().Type)
	assert.Equal(t, []string{typ}, itemSchema.Schema().OneOf[1].Schema().Type)
	if len(format) > 0 {
		assert.Equal(t, format[0], itemSchema.Schema().OneOf[1].Schema().Format)
	}
}
