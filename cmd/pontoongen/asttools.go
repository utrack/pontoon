package main

import (
	"go/ast"
	"go/token"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

func astFindFile(pkg *packages.Package, pos token.Pos) (*ast.File, error) {

	afiles := pkg.Syntax

	selFile := pkg.Fset.File(pos)

	var af *ast.File

	for _, sf := range afiles {
		asf := pkg.Fset.File(sf.Pos())
		if asf == selFile {
			af = sf
			break
		}
	}
	if af == nil {
		return nil, errors.Errorf("file '%v' not found in fset", selFile.Name())
	}
	return af, nil
}
