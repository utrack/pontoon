package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// getHandlerNames scans RegisterHTTP's function body and
// returns registered ops/paths/handler function AST.
func (b builder) getHandlerNames(sel *types.Func) ([]hdlPathPtr, error) {
	af, err := b.astFindFile(sel.Pos())
	if err != nil {
		return nil, err
	}

	srcPkg := sel.Pkg()

	reg, exact := astutil.PathEnclosingInterval(af, sel.Scope().Pos(), sel.Scope().End())
	if !exact {
		return nil, errors.New("cannot find exact func path")
	}

	funcDecl := reg[0].(*ast.FuncDecl)
	funcBody := funcDecl.Body
	funcRouterParam := funcDecl.Type.Params.List[0]

	vis := &visRegHTTP{
		muxDecl: funcRouterParam,
		pkg:     srcPkg,
		pkgReg:  b.pkg,
	}

	ast.Walk(vis, funcBody)

	return vis.hits, nil
}

type visRegHTTP struct {
	muxDecl *ast.Field
	hits    []hdlPathPtr
	pkg     *types.Package
	pkgReg  *packages.Package
}

func (vr *visRegHTTP) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	cv, ok := node.(*ast.CallExpr)
	if !ok {
		return vr
	}

	se, ok := cv.Fun.(*ast.SelectorExpr)
	if !ok {
		return vr
	}

	muxIdent, ok := se.X.(*ast.Ident)
	if !ok {
		return vr
	}

	if muxIdent.Obj.Decl != vr.muxDecl {
		return vr
	}

	argOp,
		argPath,
		argHandlerFunc :=
		vr.litFromExpr(cv.Args[0]),
		vr.litFromExpr(cv.Args[1]),
		cv.Args[2].(*ast.SelectorExpr)

	vr.hits = append(vr.hits, hdlPathPtr{
		op:   strings.Trim(argOp.Value, `"`),
		path: strings.Trim(argPath.Value, `"`),
		fn:   argHandlerFunc})

	return nil
}

func (vr *visRegHTTP) litFromExpr(ex ast.Node) *ast.BasicLit {
	switch v := ex.(type) {
	case *ast.BasicLit:
		return v
	case *ast.Ident:
		vv := v.Obj.Decl.(*ast.ValueSpec).Values[0]
		return vr.litFromExpr(vv)
	case *ast.SelectorExpr: // usually reference from another package
		pkg := findImportedPackage(vr.pkg, v.X.(*ast.Ident).Name)
		obj := pkg.Scope().Lookup(v.Sel.Name)

		f, _ := astFindFile(vr.pkgReg.Imports[pkg.Path()], obj.Pos())

		ecl, _ := astutil.PathEnclosingInterval(f, obj.Pos()-1, obj.Pos())

		specs := ecl[0].(*ast.GenDecl).Specs
		for i, s := range specs {
			specName := s.(*ast.ValueSpec).Names[0].Name
			if specName == v.Sel.Name {
				return vr.litFromExpr(specs[i])
			}
		}

		return vr.litFromExpr(ecl[0].(*ast.GenDecl).Specs[0])
	case *ast.ValueSpec:
		return vr.litFromExpr(v.Names[0])
	default:
		panic(fmt.Sprintf("litFromExpr: cannot convert %v (%v) to BasicLit", ex, reflect.TypeOf(ex).String()))
	}
}

func findImportedPackage(pkg *types.Package, name string) *types.Package {
	for _, p := range pkg.Imports() {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
