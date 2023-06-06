package main

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

func genOpenAPI(ss []serviceDesc, pkgName string) ([]byte, error) {

	paths := openapi3.Paths{}

	for _, s := range ss {
		for _, h := range s.handlers {
			p := paths[h.path]
			if p == nil {
				p = &openapi3.PathItem{}
				paths[h.path] = p
			}

			op := openapi3.NewOperation()
			op.Description = h.description
			op.Tags = []string{s.name}

			if h.inout.inType != nil {
				err := genInSchema(h.inout.inType, op)
				if err != nil {
					return nil, errors.Wrap(err, "generating input schema")
				}
			}

			out, err := genRefOut(h.inout.outType)
			if err != nil {
				return nil, errors.Wrap(err, "generating output schema")
			}

			rsp := openapi3.NewResponse()
			rsp = rsp.WithDescription("success")
			rsp.Content = openapi3.NewContentWithJSONSchemaRef(out)
			op.AddResponse(200, rsp)
			p.SetOperation(h.op, op)
		}
	}

	comp := openapi3.NewComponents()
	comp.Schemas = openapi3.Schemas{}
	for d, t := range cacheSchemaRefs {
		if t.Value == nil || t.Value.Type != "object" {
			continue
		}

		comp.Schemas[d.name] = openapi3.NewSchemaRef("", t.Value)
	}

	root := openapi3.T{}
	root.Info = &openapi3.Info{
		Title:   pkgName,
		Version: "1-autogen",
	}

	root.OpenAPI = "3.1.0"
	root.Components = comp
	root.Paths = paths

	//err := root.Validate(context.Background())
	ret, err := json.MarshalIndent(&root, "  ", "  ")
	if err != nil {
		panic(fmt.Sprintf("error marshalling openapi spec: %s", err))
	}
	cacheSchemaRefs = map[*typeDesc]*openapi3.SchemaRef{}
	return ret, nil
}

var cacheSchemaRefs = map[*typeDesc]*openapi3.SchemaRef{}

func genInSchema(t *typeDesc, sc *openapi3.Operation) error {
	// Dereference pointers in input parameters to get the actual type
	if t.isStruct == nil && t.isPtr != nil {
		t = t.isPtr
	}

	for _, f := range t.isStruct.embeds {
		if f.t.id == "github.com/ggicci/httpin.JSONBody" {
			rs, err := genRefFieldStruct(t)
			if err != nil {
				return err
			}

			body := openapi3.NewRequestBody().WithJSONSchemaRef(rs)
			sc.RequestBody = &openapi3.RequestBodyRef{
				Value: body,
			}
			continue
		}
		err := genInSchema(f.t, sc)
		if err != nil {
			return err
		}
	}

	for _, f := range t.isStruct.fields {
		props := genInProps(f.tags)
		if props == nil {
			continue
		}

		fs, err := genFieldSchema(f)
		if err != nil {
			return err
		}

		fmt.Println(props)
		doc := docFromComment(f.name, props.name, f.doc)
		switch props.location {
		case "body":
			body := openapi3.NewRequestBody().WithJSONSchemaRef(fs).WithDescription(doc)
			sc.RequestBody = &openapi3.RequestBodyRef{
				Value: body,
			}
		case "query":
			q := openapi3.NewQueryParameter(props.name).
				WithSchema(fs.Value).
				WithRequired(props.required).
				WithDescription(doc)

			sc.AddParameter(q)
		case "header":
			q := openapi3.NewHeaderParameter(props.name).
				WithSchema(fs.Value).
				WithRequired(props.required).
				WithDescription(doc)
			sc.AddParameter(q)
		case "path":
			q := openapi3.NewPathParameter(props.name).
				WithSchema(fs.Value).
				WithRequired(props.required).
				WithDescription(doc)
			sc.AddParameter(q)
		default:
			return errors.Errorf("unknown in source type '%v' for field '%v'", props.location, f.name)
		}
	}
	return nil
}

func genFieldSchema(f descField) (*openapi3.SchemaRef, error) {

	if f.t.isScalar {
		ret, err := genRefFieldScalar(f.t)
		if err != nil {
			return nil, err
		}
		ret.Value.Description = docFromComment(f.name, "", f.doc)
		return ret, nil
	}
	if f.t.isStruct != nil {
		ref, err := genRefFieldStruct(f.t)
		if err != nil {
			return nil, err
		}
		return openapi3.NewSchemaRef(ref.Ref, nil), nil
	}
	if f.t.isMap != nil {
		return genRefFieldMap(f.t)
	}
	if f.t.isSlice != nil {
		return genRefFieldSlice(f.t)
	}
	if f.t.isPtr != nil {
		f.t = f.t.isPtr
		val, err := genFieldSchema(f)
		if err != nil {
			return nil, err
		}
		if f.t.isScalar {
			val.Ref = ""
			val.Value.Nullable = true
			return val, nil
		}

		nulltype := openapi3.NewSchema()
		nulltype.Type = "null"

		ret := openapi3.NewSchema()
		ret.AnyOf = append(ret.AnyOf,
			openapi3.NewSchemaRef("", nulltype),
			val)
		return openapi3.NewSchemaRef("", ret), nil
	}
	if f.t.isSpecial != 0 {
		return genRefFieldSpecial(f.t)
	}
	if f.t.isAny {
		return genRefFieldAny(f.t)
	}
	panic(fmt.Sprintf("failed to generate field schema: %+v", f.t))
}

func genRefOut(t *typeDesc) (*openapi3.SchemaRef, error) {
	if t == nil {
		return nil, nil
	}

	if t.isAny {
		return genRefFieldAny(t)
	}
	if t.isSlice != nil {
		return genRefFieldSlice(t)
	}
	if t.isMap != nil {
		return genRefFieldMap(t)
	}
	if t.isScalar {
		return genRefFieldScalar(t)
	}
	return genRefFieldStruct(t)
}

func genRefFieldStruct(t *typeDesc) (*openapi3.SchemaRef, error) {
	if t == nil {
		return nil, nil
	}

	if t.isStruct == nil {
		panic(fmt.Sprintf("generating ref for struct field but t.isStruct is false - t: %+v", *t))
	}
	if ref, ok := cacheSchemaRefs[t]; ok {
		return ref, nil
	}

	sc := openapi3.NewSchema()
	sc.Properties = openapi3.Schemas{}
	ref := "#/components/schemas/" + t.name
	ret := openapi3.NewSchemaRef(ref, sc)
	cacheSchemaRefs[t] = ret

	sc.Type = "object"
	sc.Description = docFromComment(t.name, "", t.doc)

	for _, e := range t.isStruct.embeds {
		if e.t.id == "github.com/ggicci/httpin.JSONBody" {
			continue
		}

		ref, err := genFieldSchema(e)
		if err != nil {
			return nil, errors.Wrapf(err, "processing embedded field '%v'", e.name)
		}
		sc.AllOf = append(sc.AllOf, openapi3.NewSchemaRef(ref.Ref, nil))
	}

	for _, f := range t.isStruct.fields {
		ref, err := genFieldSchema(f)
		if err != nil {
			return nil, errors.Wrapf(err, "processing embedded field '%v'", f.name)
		}
		fname := genFieldName(f.name, f.tags)
		if fname == "-" {
			continue
		}
		sc.Properties[fname] = ref
	}
	return ret, nil
}

type inProps struct {
	name     string
	location string
	required bool
}

func genInProps(tags string) *inProps {
	tags = strings.Trim(tags, "`")
	tag := reflect.StructTag(tags)
	tagval := tag.Get("in")
	if len(tagval) == 0 || tagval == "-" {
		return nil
	}

	spl := strings.Split(tagval, ";")

	ret := &inProps{}
	for _, s := range spl {
		if s == "required" {
			ret.required = true
			continue
		}
		src2name := strings.SplitN(s, "=", 2)
		ret.location = src2name[0]
		ret.name = strings.Split(src2name[1], ",")[0]
		break
	}
	return ret
}

func genFieldName(name, tags string) string {
	tags = strings.Trim(tags, "`")
	tag := reflect.StructTag(tags)
	ret := tag.Get("json")
	ret = strings.TrimSuffix(ret, ",omitempty")
	if ret != "" {
		return ret
	}
	return name
}

func genRefFieldAny(t *typeDesc) (*openapi3.SchemaRef, error) {
	if !t.isAny {
		panic(fmt.Sprintf("generating ref for any, but t.isAny is false - t: %+v", t))
	}

	sc := openapi3.NewSchema()

	return openapi3.NewSchemaRef("", sc), nil
}

func genRefFieldScalar(t *typeDesc) (*openapi3.SchemaRef, error) {
	if !t.isScalar {
		panic(fmt.Sprintf("generating ref for scalar field, but t.isScalar is false - t: %+v", t))
	}

	sc := openapi3.NewSchema()

	switch t.name {
	case "int32":
		sc.Type = "integer"
		sc.Format = "int32"
	case "int", "int64":
		sc.Type = "integer"
		sc.Format = "int64"
	case "int16":
		sc.Type = "integer"
		sc.Format = "int16"
	case "float32":
		sc.Type = "number"
		sc.Format = "float"
	case "float", "float64":
		sc.Type = "number"
		sc.Format = "double"
	case "string":
		sc.Type = "string"
	case "uint", "uint64":
		sc.Type = "integer"
		sc.Format = "int64"
		sc = sc.WithMin(0)
	case "uint32":
		sc.Type = "integer"
		sc.Format = "int32"
		sc = sc.WithMin(0)
	case "uint8":
		sc.Type = "integer"
		sc = sc.WithMin(0).WithMax(math.MaxUint8)
	case "bool":
		sc.Type = "boolean"
	default:
		return nil, errors.Errorf("unknown scalar field type '%v'", t.name)
	}
	return openapi3.NewSchemaRef("", sc), nil
}

func genRefFieldSlice(t *typeDesc) (*openapi3.SchemaRef, error) {
	if t == nil {
		return nil, nil
	}

	if t.isSlice == nil {
		panic(fmt.Sprintf("generating ref for slice, but t.isSlice is nil, t: %+v", *t))
	}

	// Represent []byte as string with byte format
	if t.isSlice.t.name == "byte" {
		sc := openapi3.NewSchema()
		sc.Type = "string"
		sc.Format = "byte"
		return openapi3.NewSchemaRef("", sc), nil
	}

	val, err := genFieldSchema(descField{t: t.isSlice.t})
	if err != nil {
		return nil, errors.Wrap(err, "creating slice value ref")
	}

	ret := openapi3.NewSchema()
	ret.Type = "array"
	ret.Items = val
	ret.Nullable = true
	return openapi3.NewSchemaRef("", ret), nil
}

func genRefFieldMap(t *typeDesc) (*openapi3.SchemaRef, error) {
	if t == nil {
		return nil, nil
	}

	if t.isMap == nil {
		panic(fmt.Sprintf("generating ref for map, but t.isMap is false - t: %+v", t))
	}

	if !t.isMap.key.isScalar {
		return nil, errors.New("non-scalar keys in maps are not allowed")
	}
	val, err := genFieldSchema(descField{t: t.isMap.value})
	if err != nil {
		return nil, err
	}

	ret := openapi3.NewSchema()
	ret.Type = "object"
	ret.AdditionalProperties = val
	return openapi3.NewSchemaRef("", ret), nil
}

func genRefFieldSpecial(t *typeDesc) (*openapi3.SchemaRef, error) {
	switch t.isSpecial {
	case specialTypeTime:
		sc := openapi3.NewSchema()
		sc.Type = "string"
		sc.Format = "date-time"
		return openapi3.NewSchemaRef("", sc), nil
	default:
		panic(fmt.Sprintf("unsupported special field - t: %+v", t))
	}
}

func docFromComment(goLongName string, jsonTag string, comment string) string {

	goName := goLongName
	// foo.Bar -> Bar
	if idx := strings.LastIndex(goLongName, "."); idx > -1 {
		goName = goName[idx+1:]
	}

	// remove heading FieldName needed by Go specs
	//
	// FooField is a foo field -> is a foo field
	comment = strings.TrimPrefix(comment, goName)

	comment = strings.Trim(comment, "\n\r \t")

	comment = strings.TrimPrefix(comment, "is ")

	// replace any other goName occurences with jsonTag if it's there
	if jsonTag != "" {
		comment = strings.ReplaceAll(comment, goName, jsonTag)
	}

	// capitalize first letter
	if len(comment) > 0 {
		r := []rune(comment)
		r[0] = unicode.ToUpper(r[0])
		comment = string(r)
	}
	return comment
}
