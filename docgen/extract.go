package docgen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// Extractor extracts documentation from Go source code.
type Extractor struct {
	fset *token.FileSet
	pkg  *packages.Package

	pkgSet map[string]*packages.Package
}

// NewExtractor creates a new documentation extractor.
func NewExtractor(pkg *packages.Package, pkgSet map[string]*packages.Package) *Extractor {
	return &Extractor{
		fset:   pkg.Fset,
		pkg:    pkg,
		pkgSet: pkgSet,
	}
}

// ExtractType extracts documentation from any type, creating a TypeDoc.
func (e *Extractor) ExtractType(t *types.Named) (map[string]TypeDoc, error) {
	// Extract type information
	tt := make(map[string]TypeDoc)
	err := e.extractType(tt, t)
	return tt, errors.Wrap(err, "failed to extract type info")
}

// extractType recursively extracts type documentation
func (e *Extractor) extractType(typeDocs map[string]TypeDoc, t types.Type) error {
	switch t := t.(type) {
	case *types.Named:
		if types.IsInterface(t) {
			if t.String() == "error" {
				return nil
			}

			// TODO do smth with interfaces
		}
		typeName := typeString(t)
		if _, exists := typeDocs[typeName]; exists {
			return nil // Already processed
		}
		// prevent infinite recursion
		typeDocs[typeName] = TypeDoc{}

		// Get type position and AST node
		pos := e.fset.Position(t.Obj().Pos())

		tokFile := e.fset.File(t.Obj().Pos())
		if tokFile == nil {
			return errors.Errorf("file %s not found in the fileset", pos.Filename)
		}

		// Find the type's file
		var f *ast.File

		pkg := e.pkgSet[t.Obj().Pkg().Path()]
		for _, file := range pkg.Syntax {
			if e.fset.Position(file.Pos()).Filename == pos.Filename {
				f = file
				break
			}
		}
		if f == nil {
			fmt.Println("WARN f is nil - ret, looking for '", t.Obj().Pkg().Path(), " - ", t.Obj().Name(), " current is ", e.pkg.PkgPath)
			// Type is from another package, skip detailed extraction
			typeDocs[typeName] = TypeDoc{
				Name:    t.Obj().Name(),
				Package: t.Obj().Pkg().Path(),
				File:    pos.Filename,
				Line:    pos.Line,
			}
			return nil
		}

		// Find type declaration
		path, _ := astutil.PathEnclosingInterval(f, t.Obj().Pos(), t.Obj().Pos())
		if len(path) == 0 {
			return errors.Errorf("%s: type declaration not found", pos)
		}

		// Extract type documentation
		var typeSpec *ast.TypeSpec
		var genDecl *ast.GenDecl
		for _, node := range path {
			if ts, ok := node.(*ast.TypeSpec); ok {
				typeSpec = ts
			}
			if gd, ok := node.(*ast.GenDecl); ok {
				genDecl = gd
			}
		}
		if typeSpec == nil {
			return errors.Errorf("%s: not a type declaration", pos)
		}

		// Create type doc
		typeDoc := TypeDoc{
			Name:    t.Obj().Name(),
			Package: t.Obj().Pkg().Path(),
			File:    pos.Filename,
			Line:    pos.Line,
		}
		if pkg.Module == nil {
			typeDocs[typeName] = typeDoc
			return nil
		}

		// Get comments
		var comments []string
		if genDecl != nil && genDecl.Doc != nil {
			comments = append(comments, genDecl.Doc.Text())
		}
		if typeSpec.Doc != nil {
			comments = append(comments, typeSpec.Doc.Text())
		}
		if len(comments) > 0 {
			typeDoc.Comment = strings.Join(comments, "\n")
		}

		// Extract functions and their types
		for i := 0; i < t.NumMethods(); i++ {
			method := t.Method(i)

			// get method function's docs - name pos etc
			funDoc, err := e.extractFunction(f, method)
			if err != nil {
				return errors.Wrapf(err, "extracting method signature from '%v'", method.Name())
			}
			if funDoc == nil {
				continue
			}
			typeDoc.Methods = append(typeDoc.Methods, *funDoc)

			sig, ok := method.Type().(*types.Signature)
			if !ok {
				return nil
			}

			// extract docs for every parameter
			for i := 0; i < sig.Params().Len(); i++ {
				param := sig.Params().At(i)
				inputType := param.Type()
				if inputType == nil {
					continue
				}
				err := e.extractType(typeDocs, inputType)
				if err != nil {
					return errors.Wrapf(err, "extracting input type for method %s", method.Name())
				}
			}

			// extract docs for every result
			for i := 0; i < sig.Results().Len(); i++ {
				result := sig.Results().At(i)
				outputType := result.Type()
				if outputType == nil {
					continue
				}
				err := e.extractType(typeDocs, outputType)
				if err != nil {
					return errors.Wrapf(err, "extracting output type for method %s", method.Name())
				}
			}

			typeDoc.Methods[len(typeDoc.Methods)-1] = *funDoc
		}

		// Extract fields for structs
		if st, ok := t.Underlying().(*types.Struct); ok {
			typeDoc.IsStruct = true
			for i := 0; i < st.NumFields(); i++ {
				field := st.Field(i)

				file := e.fset.Position(field.Pos())

				// Get field AST node for comments and tags
				fieldPath, _ := astutil.PathEnclosingInterval(f, field.Pos(), field.Pos())
				var fieldNode *ast.Field
				for _, node := range fieldPath {
					if f, ok := node.(*ast.Field); ok {
						fieldNode = f
						break
					}
				}

				_, isPointer := field.Type().(*types.Pointer)

				if field.Anonymous() && !field.Embedded() {
					return errors.Errorf("%v:%v: anonymous fields not supported", file.Filename, file.Line)
				}

				fieldDoc := FieldDoc{
					Name:       field.Name(),
					Type:       typeString(field.Type()),
					FilePath:   file.Filename,
					Line:       file.Line,
					Nullable:   isPointer,
					IsEmbedded: field.Embedded(),
				}

				if fieldNode != nil {
					if fieldNode.Doc != nil {
						fieldDoc.Comment = fieldNode.Doc.Text()
					}
					if fieldNode.Tag != nil {
						tv := fieldNode.Tag.Value
						tv = strings.TrimFunc(tv, func(r rune) bool {
							return r == '`'
						})
						fieldDoc.Tags = tv
					}
				}

				if mt, ok := field.Type().(*types.Map); ok {
					fieldDoc.IsMap = &FieldMapDoc{
						TypeKey:   typeString(mt.Key()),
						TypeValue: typeString(mt.Elem()),
					}
				}

				if at, ok := field.Type().(*types.Slice); ok {
					fieldDoc.IsArray = &FieldArrayDoc{
						Type: typeString(at.Elem()),
					}
				}

				typeDoc.Fields = append(typeDoc.Fields, fieldDoc)

				// Recursively process field type
				if err := e.extractType(typeDocs, field.Type()); err != nil {
					return errors.Wrapf(err, "extracting field %s type", field.Name())
				}
			}
		}

		typeDocs[typeName] = typeDoc

		// Process underlying type
		if err := e.extractType(typeDocs, t.Underlying()); err != nil {
			return errors.Wrapf(err, "extracting underlying type of %s", typeName)
		}

	case *types.Struct:
		// Anonymous struct
		for i := 0; i < t.NumFields(); i++ {
			if err := e.extractType(typeDocs, t.Field(i).Type()); err != nil {
				return errors.Wrapf(err, "extracting anonymous struct field %d type", i)
			}
		}

	case *types.Slice:
		return e.extractType(typeDocs, t.Elem())
	case *types.Array:
		return e.extractType(typeDocs, t.Elem())
	case *types.Map:
		if err := e.extractType(typeDocs, t.Key()); err != nil {
			return errors.Wrap(err, "extracting map key type")
		}
		return e.extractType(typeDocs, t.Elem())

	case *types.Pointer:
		return e.extractType(typeDocs, t.Elem())

	case *types.Interface:
		// Skip interface extraction for now
		return nil
	}

	return nil
}

func typeString(t types.Type) string {
	if v, ok := t.(*types.Pointer); ok {
		t = v.Elem()
	}
	return types.TypeString(t, func(pkg *types.Package) string {
		return pkg.Path()
	})
}

func (e *Extractor) extractFunction(f *ast.File, fun *types.Func) (*FunctionDoc, error) {

	sig, ok := fun.Type().(*types.Signature)
	if !ok {
		return nil, nil
	}

	// Get method position
	funcPos := e.fset.Position(fun.Pos())

	// Find method declaration
	funcPath, _ := astutil.PathEnclosingInterval(f, fun.Pos(), fun.Pos())
	if len(funcPath) == 0 {
		return nil, nil
	}

	// Get method documentation
	var funcComment string
	for _, node := range funcPath {
		if fd, ok := node.(*ast.FuncDecl); ok {
			if fd.Doc != nil {
				funcComment = fd.Doc.Text()
			}
			break
		}
	}

	doc := &FunctionDoc{
		Name:    fun.Name(),
		Comment: funcComment,
		File:    funcPos.Filename,
		Line:    funcPos.Line,
	}

	for i := 0; i < sig.Params().Len(); i++ {
		var inputType types.Type
		param := sig.Params().At(i)
		if types.IsInterface(param.Type()) {
			continue
		}
		if param.Type().String() == "*net/http.Request" {
			continue
		}
		inputType = param.Type()
		if inputType != nil {
			inputTypeName := typeString(inputType)
			doc.Params = append(doc.Params, FunctionParamDoc{
				Name: param.Name(),
				Type: inputTypeName,
			})
		}
	}

	for i := 0; i < sig.Results().Len(); i++ {
		var outputType types.Type
		result := sig.Results().At(i)

		outputType = result.Type()
		if outputType != nil {
			outputTypeName := typeString(outputType)
			doc.Returns = append(doc.Returns, FunctionParamDoc{
				Name: result.Name(),
				Type: outputTypeName,
			})
		}
	}

	return doc, nil
}
