package main

import (
	"go/ast"
	"go/token"

	"github.com/pkg/errors"
)

func (b builder) astFindFile(pos token.Pos) (*ast.File, error) {

	afiles := b.pkg.Syntax

	selFile := b.pkg.Fset.File(pos)

	var af *ast.File

	for _, sf := range afiles {
		asf := b.pkg.Fset.File(sf.Pos())
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
