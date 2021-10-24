package main

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
)

// getHandlerNames scans RegisterHTTP's function body and
// returns registered ops/paths/handler function AST.
func (b builder) getHandlerNames(sel *types.Func) ([]hdlPathPtr, error) {
	af, err := b.astFindFile(sel.Pos())
	if err != nil {
		return nil, err
	}

	reg, exact := astutil.PathEnclosingInterval(af, sel.Scope().Pos(), sel.Scope().End())
	if !exact {
		return nil, errors.New("cannot find exact func path")
	}

	funcBody := reg[0].(*ast.BlockStmt)
	funcRouterParam := reg[1].(*ast.FuncDecl).Type.Params.List[0]

	vis := &visRegHTTP{muxDecl: funcRouterParam}

	ast.Walk(vis, funcBody)

	return vis.hits, nil
}

type visRegHTTP struct {
	muxDecl *ast.Field
	hits    []hdlPathPtr
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
		cv.Args[0].(*ast.BasicLit),
		cv.Args[1].(*ast.BasicLit),
		cv.Args[2].(*ast.SelectorExpr)

	vr.hits = append(vr.hits, hdlPathPtr{
		op:   strings.Trim(argOp.Value, `"`),
		path: strings.Trim(argPath.Value, `"`),
		fn:   argHandlerFunc})

	return nil
}
