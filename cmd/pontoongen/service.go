package main

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type builder struct {
	pkg *packages.Package

	muxType *types.Interface
}

func (b builder) Service(ms *types.MethodSet, hs *types.Named, fset *token.FileSet) (*serviceDesc, error) {
	hdlRegFunc := ms.Lookup(b.pkg.Types, "RegisterHTTP").Obj().(*types.Func)

	hpp, err := b.getHandlerNames(hdlRegFunc)
	if err != nil {
		return nil, errors.Wrap(err, "cannot retrieve handler names")
	}
	tokfile := b.pkg.Fset.File(hs.Obj().Pos())
	tokfile.Name()

	f, err := b.astFindFile(hs.Obj().Pos())
	if err != nil {
		return nil, errors.Wrap(err, "cannot get file for service")
	}

	// extract comments for service
	reg, _ := astutil.PathEnclosingInterval(f, hs.Obj().Pos()-1, hs.Obj().Pos())

	doc := reg[0].(*ast.GenDecl).Doc

	hd := []hdlDesc{}

	for _, hp := range hpp {
		fnDesc, err := b.getHandleDesc(hp.fn.Sel, ms)
		if err != nil {
			return nil, errors.Wrapf(err, "when parsing '%v' '%v' of '%v'", hp.op, hp.path, hs.String())
		}

		hd = append(hd, hdlDesc{
			op:    hp.op,
			path:  hp.path,
			inout: *fnDesc,
		})
	}
	ret := serviceDesc{
		name:              hs.Obj().Pkg().Name() + "." + hs.Obj().Name(),
		handlers:          hd,
		pkg:               hs.Obj().Pkg().Path(),
		filename:          tokfile.Name(),
		serviceStructName: hs.Obj().Name(),
	}
	if doc != nil {
		ret.doc = doc.Text()
	}

	return &ret, nil
}

type hdlPathPtr struct {
	op   string
	path string
	fn   *ast.SelectorExpr
}
