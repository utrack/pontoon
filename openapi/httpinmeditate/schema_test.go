package httpinmeditate

import (
	"encoding"
	"reflect"
	"testing"
	"time"

	"github.com/ggicci/httpin"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	oext "github.com/utrack/pontoon/v2/openapi/pontoonext"
)

func init() {
	//debugEnabled = true
}

type Address struct {
	Street string `in:"form=street"`
	City   string `in:"form=city"`
}

type User struct {
	Address                     // embedded struct
	ID        int64             `in:"path=id"`
	Name      *string           `in:"form=name"`
	Age       int               `in:"query=age"`
	Tags      []string          `in:"form=tags"`
	Metadata  map[string]string `in:"json=metadata"`
	CreatedAt time.Time         `in:"header=created-at"`
	UpdatedAt *time.Time        `in:"header=updated-at"`
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
	Value    string `in:"form=value"`
	Parent   *Node  `in:"form=parent"`
	Children []Node `in:"form=children"`
}

type BaseA struct {
	FieldA string `in:"form=field_a"`
}

type BaseB struct {
	FieldB string `in:"form=field_b"`
}

type MultiEmbed struct {
	BaseA
	*BaseB
	FieldC string `in:"form=field_c"`
}

type ComplexTypes_SlicePtr struct {
	NestedPtrSlice []*BaseA `in:"form=nested_ptr_slice"`
}

type ComplexTypes_StructMap struct {
	StructMap map[string]BaseA `in:"form=struct_map"`
}

type ComplexTypes struct {
	String string  `in:"form=string"`
	Int    int     `in:"form=int"`
	Float  float64 `in:"form=float"`
	Bool   bool    `in:"form=bool"`
	Bytes  []byte  `in:"form=bytes"`

	CustomStr  CustomString    `in:"form=custom_str"`
	CustomInt  CustomInt       `in:"form=custom_int"`
	CustomTime CustomTime      `in:"form=custom_time"`
	Marshaler  MarshalableType `in:"form=marshaler"`

	StringPtr *string  `in:"form=string_ptr"`
	IntPtr    *int     `in:"form=int_ptr"`
	FloatPtr  *float64 `in:"form=float_ptr"`
	BoolPtr   *bool    `in:"form=bool_ptr"`

	StringArray  [3]string `in:"form=string_array"`
	IntSlice     []int     `in:"form=int_slice"`
	Float64Slice []float64 `in:"form=float_slice"`

	StringMap map[string]string `in:"form=string_map"`
	IntMap    map[string]int    `in:"form=int_map"`
	StructMap map[string]BaseA  `in:"form=struct_map"`

	Nested         BaseA    `in:"form=nested"`
	NestedPtr      *BaseA   `in:"form=nested_ptr"`
	NestedSlice    []BaseA  `in:"form=nested_slice"`
	NestedPtrSlice []*BaseA `in:"form=nested_ptr_slice"`

	Time         time.Time    `in:"form=time"`
	TimePtr      *time.Time   `in:"form=time_ptr"`
	TimeSlice    []time.Time  `in:"form=time_slice"`
	TimePtrSlice []*time.Time `in:"form=time_ptr_slice"`

	EmptyStruct struct{} `in:"form=empty_struct"`
	// AnonStruct  struct {
	// 	Field string `in:"form=field"`
	// } `in:"form=anon_struct"`
}

type structWithFile struct {
	File *httpin.File `in:"form=file"`
}

func TestGenerator_SchemaWithFile(t *testing.T) {
	g := NewGenerator()

	//debugEnabled = true
	schema, err := g.generateSchema(reflect.TypeOf(structWithFile{}))
	require.NoError(t, err)
	require.NotNil(t, schema)

	// Convert schema to map for easier testing
	doc := schema.Schema()
	assert.Equal(t, []string{"object"}, doc.Type)

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
	assert.Equal(t, "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.Address", addrReference)

	// Check regular fields
	regularSchema := doc.AllOf[1].Schema()
	assert.Equal(t, []string{"object"}, regularSchema.Type)
	require.NotNil(t, regularSchema.Properties)
	props := regularSchema.Properties

	// Check ID field
	idSchema := props.GetOrZero("id").Schema()
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
				Field interface{} `in:"form=field"`
			}{}),
			wantErr:  true,
			errMatch: "type interface {} not supported",
		},
		{
			name: "map with non-string key",
			typ: reflect.TypeOf(struct {
				Field map[int]string `in:"form=field"`
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
	assertArrayNullableRefType(t, props, "nested_ptr_slice", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")
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

	assertMapRefType(t, props, "struct_map", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")
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
	assertMapRefType(t, props, "struct_map", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")

	// Test nested structs
	assertRefType(t, props, "nested", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")
	assertNullableRefType(t, props, "nested_ptr", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")
	assertArrayRefType(t, props, "nested_slice", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")
	assertArrayNullableRefType(t, props, "nested_ptr_slice", "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA")

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
	_, err = doc.Render()
	require.Nil(t, err)

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
	assert.Equal(t, "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.Node", parentSchema.OneOf[1].Schema().AllOf[0].GetReference())

	// Check Children field (recursive slice)
	childrenSchema := props.GetOrZero("children").Schema()
	assert.Equal(t, []string{"array"}, childrenSchema.Type)
	assert.Equal(t, "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.Node", childrenSchema.Items.A.Schema().AllOf[0].GetReference())
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
	assert.Equal(t, "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseA", baseARef)

	// Check BaseB fields (pointer)
	baseBPtrSchema := doc.AllOf[1].Schema()
	require.Len(t, baseBPtrSchema.OneOf, 2)
	assert.Equal(t, []string{"null"}, baseBPtrSchema.OneOf[0].Schema().Type)
	assert.Equal(t, "#/components/schemas/github.com_utrack_pontoon_v2_openapi_httpinmeditate.BaseB", baseBPtrSchema.OneOf[1].GetReference())

	// Check own fields
	ownSchema := doc.AllOf[2].Schema()
	assert.Equal(t, []string{"object"}, ownSchema.Type)
	assertSchemaType(t, ownSchema.Properties, "field_c", "string")
}

func TestGenerator_GenerateSchema_HttpinTags(t *testing.T) {
	type TestStruct struct {
		FormField   string            `in:"form=form_field"`
		QueryField  *string           `in:"query=query_field"`
		HeaderField []int             `in:"header=header-field"`
		CookieField bool              `in:"cookie=cookie_field"`
		PathField   int64             `in:"path=path_field"`
		JSONField   map[string]string `in:"json=json_field"`
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
	FieldJSON StructWithJSONField `in:"body=json"`
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

	ref := "github.com_utrack_pontoon_v2_openapi_httpinmeditate.StructWithJSONField"
	refString := "#/components/schemas/" + ref

	// workaround for OAPI 3.0-style embeddings
	require.True(t, len(jsonField.Schema().AllOf) == 1)
	require.Equal(t, refString, jsonField.Schema().AllOf[0].GetReference())

	dict := g.Components()
	nestedStructDesc, ok := dict.Schemas.Get(ref)
	require.True(t, ok)

	require.Equal(t, []string{"object"}, nestedStructDesc.Schema().Type)

	sch := nestedStructDesc.Schema()
	require.Equal(t, sch.Properties.Len(), 3)
}

type ParamModel struct {
	ID      int    `in:"path=id"`
	Query   string `in:"query=q"`
	Header  string `in:"header=x-custom"`
	Cookie  string `in:"cookie=session"`
	User    User   `in:"body=json"`
	Nested  ParamModelNested
	Ignored string
}

type ParamModelNested struct {
	Field1 string `in:"path=field1"`
	Field2 string `in:"query=field2"`
}

func TestGenerator_GenerateOperationModel(t *testing.T) {
	g := NewGenerator()

	schema, params, err := g.GenerateOperationRequestParams(reflect.TypeOf(ParamModel{}))
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
	_, err = doc.Render()
	require.Nil(t, err)
	// Check parameters
	require.Len(t, params, 6, "should include all parameters including nested structs")

	// Verify each parameter
	for _, p := range params {

		switch p.Name {
		case "id":
			assert.Equal(t, "path", p.In)
			assert.Equal(t, []string{"integer"}, p.Schema.Schema().Type)
			annot, ok := p.Extensions.Get(oext.ExtGoFieldName)
			assert.True(t, ok)
			assert.NotNil(t, annot)

			goType, ok := oext.GetGoTypeInfo(p.Extensions)
			require.True(t, ok)
			require.NotEmpty(t, goType)
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
	require.Equal(t, ref, schema.AdditionalProperties.A.Schema().AllOf[0].GetReference())
}

func assertRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	prop := props.GetOrZero(field)
	require.NotNil(t, prop)
	refSchema := props.GetOrZero(field).Schema().AllOf[0].GetReference()
	assert.Equal(t, ref, refSchema)
}

func assertNullableRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	schema := props.GetOrZero(field).Schema()
	require.NotNil(t, schema.OneOf)
	assert.Equal(t, []string{"null"}, schema.OneOf[0].Schema().Type)
	assert.Equal(t, ref, schema.OneOf[1].Schema().AllOf[0].GetReference())
}

func assertArrayRefType(t *testing.T, props *orderedmap.Map[string, *base.SchemaProxy], field string, ref string) {
	schema := props.GetOrZero(field).Schema()
	assert.Equal(t, []string{"array"}, schema.Type)
	assert.Equal(t, ref, schema.Items.A.Schema().AllOf[0].GetReference())
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
		assert.Equal(t, ref, elem.Schema().AllOf[0].GetReference())
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
