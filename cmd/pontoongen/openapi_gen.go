package main

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/ryboe/q"
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
				_, err := genRefFieldStruct(h.inout.inType)
				if err != nil {
					return nil, errors.Wrap(err, "generating input schema")
				}
			}

			out, err := genRefFieldStruct(h.inout.outType)
			if err != nil {
				return nil, errors.Wrap(err, "generating output schema")
			}
			if out != nil {
				op.Responses = openapi3.NewResponses()
				op.Responses.Default().Ref = out.Ref
			}
			p.SetOperation(h.op, op)
		}
	}

	comp := openapi3.NewComponents()
	comp.Schemas = openapi3.Schemas{}
	for _, t := range typeCache {
		if t.isStruct == nil {
			continue
		}
		sr, ok := cacheSchemaRefs[t]
		if !ok {
			panic("not found sr for " + t.name)
		}

		comp.Schemas[t.name] = openapi3.NewSchemaRef("", sr.Value)
	}

	root := openapi3.T{}
	root.Info = &openapi3.Info{
		Title:   pkgName,
		Version: "1-autogen",
	}

	root.OpenAPI = "3.0.1"
	root.Components = comp
	root.Paths = paths

	ret, err := json.MarshalIndent(&root, "  ", "  ")
	if err != nil {
		panic(err)
	}
	return ret, nil
}

var cacheSchemaRefs = map[*typeDesc]*openapi3.SchemaRef{}

func genFieldSchema(f descField) (*openapi3.SchemaRef, error) {

	if f.t.isScalar {
		ret, err := genRefFieldScalar(*f.t)
		if err != nil {
			return nil, err
		}
		ret.Value.Description = f.doc
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
		if !f.t.isMap.key.isScalar {
			return nil, errors.New("non-scalar keys in maps are not allowed")
		}
		val, err := genFieldSchema(descField{t: f.t.isMap.value})
		if err != nil {
			return nil, err
		}

		ret := openapi3.NewSchema()
		ret.Type = "object"
		ret.AdditionalProperties = val
		return openapi3.NewSchemaRef("", ret), nil
	}
	if f.t.isSlice != nil {
		val, err := genFieldSchema(descField{t: f.t.isSlice.t})
		if err != nil {
			return nil, errors.Wrap(err, "creating slice value ref")
		}

		ret := openapi3.NewSchema()
		ret.Type = "array"
		ret.Items = val
		ret.Nullable = true
		return openapi3.NewSchemaRef("", ret), nil

	}
	if f.t.isPtr != nil {
		f.t = f.t.isPtr
		return genFieldSchema(f)
		// TODO set nullable for field
	}
	if f.t.isSpecial != 0 {
		return genRefFieldSpecial(f.t)
	}
	q.Q(f.t)
	panic(fmt.Sprint(f.t))
}

func genRefFieldStruct(t *typeDesc) (*openapi3.SchemaRef, error) {
	if t == nil {
		return nil, nil
	}
	q.Q(t.name)

	if t.isStruct == nil {
		panic(fmt.Sprint(*t))
	}
	if ref, ok := cacheSchemaRefs[t]; ok {
		return ref, nil
	}

	sc := openapi3.NewSchema()
	sc.Properties = openapi3.Schemas{}

	sc.Type = "object"
	sc.Description = t.doc

	for _, e := range t.isStruct.embeds {
		ref, err := genRefFieldStruct(e)
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
		sc.Properties[fname] = ref
	}
	ret := openapi3.NewSchemaRef("#/components/schemas/"+t.name, sc)
	cacheSchemaRefs[t] = ret
	return ret, nil
}

func genFieldName(name, tags string) string {
	tags = strings.Trim(tags, "`")
	tag := reflect.StructTag(tags)
	ret := tag.Get("json")
	if ret != "" {
		return ret
	}
	return name
}

func genRefFieldScalar(t typeDesc) (*openapi3.SchemaRef, error) {
	if !t.isScalar {
		panic(t)
	}

	sc := openapi3.NewSchema()

	switch t.name {
	case "int32":
		sc.Type = "number"
		sc.Format = "int32"
	case "int", "int64":
		sc.Type = "number"
		sc.Format = "int64"
	case "float32":
		sc.Type = "number"
		sc.Format = "float"
	case "float", "float64":
		sc.Type = "number"
		sc.Format = "double"
	case "string":
		sc.Type = "string"
	case "uint", "uint64":
		sc.Type = "number"
		sc.Format = "int64"
		sc = sc.WithMin(0)
	case "uint32":
		sc.Type = "number"
		sc.Format = "int32"
		sc = sc.WithMin(0)
	case "uint8":
		sc.Type = "number"
		sc = sc.WithMin(0).WithMax(math.MaxUint8)
	case "bool":
		sc.Type = "boolean"
	default:
		return nil, errors.Errorf("unknown scalar field type '%v'", t.name)
	}
	return openapi3.NewSchemaRef("", sc), nil
}
func genRefFieldSpecial(t *typeDesc) (*openapi3.SchemaRef, error) {
	switch t.isSpecial {
	case specialTypeTime:
		sc := openapi3.NewSchema()
		sc.Type = "string"
		sc.Format = "date-time"
		return openapi3.NewSchemaRef("", sc), nil
	default:
		panic(t)
	}
}
