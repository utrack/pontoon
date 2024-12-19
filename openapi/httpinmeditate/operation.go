package httpinmeditate

import (
	"fmt"
	"reflect"
	"strings"

	oext "github.com/utrack/pontoon/openapi/pontoonext"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (g *Generator) GenerateModel(t reflect.Type) (*base.SchemaProxy, []*v3.Parameter, error) {
	schema, err := g.generateSchema(t)
	if err != nil {
		return nil, nil, err
	}

	w := &walker{
		refs:       g.Components().Schemas,
		parameters: []*v3.Parameter{},
	}
	if err := w.pullOperationParameters(schema); err != nil {
		return nil, nil, err
	}

	return w.body, w.parameters, nil
}

type walker struct {
	refs *orderedmap.Map[string, *base.SchemaProxy]

	parameters []*v3.Parameter
	body       *base.SchemaProxy
}

func (w *walker) getByRef(ref string) (*base.SchemaProxy, bool) {
	ref = strings.TrimPrefix(ref, "#/components/schemas/")
	var ok bool
	in, ok := w.refs.Get(ref)
	return in, ok
}

func (w *walker) pullOperationParameters(in *base.SchemaProxy) error {
	if in.IsReference() {
		var ok bool
		in, ok = w.refs.Get(in.GetReference())
		if !ok {
			return errors.Errorf("reference '%v' not resolved", in.GetReference())
		}
	}
	sch := in.Schema()
	if sch == nil {
		return errors.New("no schema")
	}
	for _, item := range sch.AnyOf {
		debugLog("-> next anyOf")
		if err := w.pullOperationParameters(item); err != nil {
			return err
		}
	}
	for _, item := range sch.OneOf {
		debugLog("-> next oneOf")
		if err := w.pullOperationParameters(item); err != nil {
			return err
		}
	}
	for item := range sch.Properties.ValuesFromNewest() {
		debugLog("-> next field (property)")
		if item.IsReference() {
			debugLog(" --> isReference '%v'", item.GetReference())
			err := w.pullOperationParameters(item)
			if err != nil {
				return err
			}
		}

		ext := item.Schema().Extensions

		putExtensions := orderedmap.New[string, *yaml.Node]()

		schGoType,_ := oext.GetGoTypeInfo(sch.Extensions)

		schGoType.SetTo(putExtensions)

		if ext != nil {
			if e, ok := ext.Get(oext.ExtGoFieldName); ok {
				putExtensions.Set(oext.ExtGoFieldName, e)
			}
			debugLog("-> has extensions,first: %s", ext.Newest().Key)
			in := ext.GetOrZero("x-httpin-in")
			if in != nil {

				if in.Value == "body" {
					if w.body != nil {
						return errors.New("multiple bodies found")
					}
					w.body = item
					continue
				}

				name := ext.GetOrZero(fmt.Sprintf("x-httpin-%v", in.Value))

				for k, v := range ext.FromNewest() {
					putExtensions.Set(k, &yaml.Node{
						Kind:  v.Kind,
						Tag:   v.Tag,
						Value: v.Value,
					})
				}

				w.parameters = append(w.parameters, &v3.Parameter{
					Name:       name.Value,
					In:         in.Value,
					Schema:     item,
					Extensions: putExtensions,
				})
				continue
			}
		}

		if item.Schema().SchemaTypeRef != "" {
			refed, ok := w.getByRef(item.Schema().SchemaTypeRef)
			if !ok {
				return errors.Errorf("reference '%v' not resolved", item.Schema().SchemaTypeRef)
			}
			debugLog("-> resolved SchemaTypeRef ", item.Schema().SchemaTypeRef)
			if err := w.pullOperationParameters(refed); err != nil {
				return err
			}
		}

	}
	return nil
}
