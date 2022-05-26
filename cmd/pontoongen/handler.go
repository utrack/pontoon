package main

import (
	"go/ast"
	"go/types"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
)

func (b builder) getHandleDesc(fnIdent *ast.Ident, ms *types.MethodSet) (*hdlTypesDesc, error) {
	var sel *types.Selection

	for i := 0; i < ms.Len(); i++ {
		s := ms.At(i)
		if s.Obj().Name() == fnIdent.Name {
			sel = s
			break
		}
	}
	if sel == nil {
		return nil, errors.Errorf("handler func '%v' not found", fnIdent.Name)
	}

	f, _ := b.astFindFile(sel.Obj().Pos())

	sig := sel.Type().(*types.Signature)

	var inType *typeDesc
	var outType *typeDesc

	var hasResponseWriter bool
	for i := 0; i < sig.Params().Len(); i++ {
		p := sig.Params().At(i)
		namedType := p.Type()
		if namedType.String() == "*net/http.Request" {
			continue
		}
		if namedType.String() == "net/http.ResponseWriter" {
			hasResponseWriter = true
			continue
		}

		if inType != nil {
			return nil, errors.New("handler has more than one request type")
		}

		var err error
		inType, err = b.rootStructDesc(namedType)
		if err != nil {
			return nil, errors.Wrapf(err, "converting input type '%v' to type description", namedType.String())
		}
		if inType.isAny {
			return nil, errors.Errorf("input type '%v' cannot be an interface", namedType.String())
		}
	}

	if !hasResponseWriter {
		for i := 0; i < sig.Results().Len(); i++ {
			r := sig.Results().At(i)

			t := r.Type()
			if t.String() == "error" {
				continue
			}

			if outType != nil {
				return nil, errors.New("handler has more than one response type")
			}

			var err error
			outType, err = b.rootStructDesc(t)
			if err != nil {
				return nil, errors.Wrapf(err, "converting result type '%v' to type description", t)
			}
		}
	}

	ret := &hdlTypesDesc{
		hasResponseWriter: hasResponseWriter,
		inType:            inType,
		outType:           outType,
	}
	if f != nil {
		pathdesc, _ := astutil.PathEnclosingInterval(f, sel.Obj().Pos(), sel.Obj().Pos())
		for _, n := range pathdesc {
			fd, ok := n.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fd.Name == nil ||
				fd.Name.Name != fnIdent.Name {
				continue
			}
			ret.description = fd.Doc.Text()
		}
	}
	return ret, nil
}

func (b *builder) rootStructDesc(t types.Type) (*typeDesc, error) {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	return b.getTypeDescCached(t)
}
