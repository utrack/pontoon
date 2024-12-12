package docgen

import (
	"fmt"
	"go/ast"
	"go/parser"
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
}

// NewExtractor creates a new documentation extractor.
func NewExtractor(pkg *packages.Package) *Extractor {
	return &Extractor{
		fset: pkg.Fset,
		pkg:  pkg,
	}
}

// ExtractService extracts documentation from a service type.
func (e *Extractor) ExtractService(t *types.Named) (*ServiceDoc, error) {
	pos := e.fset.Position(t.Obj().Pos())
	doc := &ServiceDoc{
		Name:    t.Obj().Name(),
		Package: e.pkg.PkgPath,
		File:    pos.Filename,
		Line:    pos.Line,
		Types:   make(map[string]TypeDoc),
	}

	// Find the AST node for the service type
	var f *ast.File
	for _, file := range e.pkg.Syntax {
		if e.fset.Position(file.Pos()).Filename == pos.Filename {
			f = file
			break
		}
	}
	if f == nil {
		return nil, errors.Errorf("file %s not found in package %s", pos.Filename, e.pkg.PkgPath)
	}

	// Find the type declaration
	path, _ := astutil.PathEnclosingInterval(f, t.Obj().Pos(), t.Obj().Pos())
	if len(path) == 0 {
		return nil, errors.Errorf("%s: type declaration not found", pos)
	}

	// Extract service-level documentation
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
		return nil, errors.Errorf("%s: not a type declaration", pos)
	}

	// Get comments from both the type spec and its parent GenDecl
	var comments []string
	if genDecl != nil && genDecl.Doc != nil {
		comments = append(comments, genDecl.Doc.Text())
	}
	if typeSpec.Doc != nil {
		comments = append(comments, typeSpec.Doc.Text())
	}
	if len(comments) > 0 {
		doc.Comments = append(doc.Comments, DocComment{
			ID: DocID{
				PkgPath:  e.pkg.PkgPath,
				TypeName: typeSpec.Name.Name,
				FilePath: pos.Filename,
				Line:     pos.Line,
			}.Hash(),
			Path:       e.pkg.PkgPath + "." + typeSpec.Name.Name,
			Comment:    strings.Join(comments, "\n"),
			SourceFile: pos.Filename,
			Line:       pos.Line,
			Type:       "service",
		})
	}

	// Extract methods
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)

		// Get method signature
		sig, ok := method.Type().(*types.Signature)
		if !ok {
			continue
		}

		// Get method position
		methodPos := e.fset.Position(method.Pos())

		// Find method declaration
		methodPath, _ := astutil.PathEnclosingInterval(f, method.Pos(), method.Pos())
		if len(methodPath) == 0 {
			continue
		}

		// Get method documentation
		var methodDoc string
		for _, node := range methodPath {
			if fd, ok := node.(*ast.FuncDecl); ok {
				if fd.Doc != nil {
					methodDoc = fd.Doc.Text()
				}
				break
			}
		}

		// Extract input/output types
		var inputType, outputType string
		if sig.Params().Len() >= 2 {
			inputType = typeString(sig.Params().At(1).Type())
			if err := e.extractType(doc.Types, sig.Params().At(1).Type()); err != nil {
				return nil, errors.Wrapf(err, "extracting input type for method %s", method.Name())
			}
		}
		if sig.Results().Len() > 0 {
			outputType = typeString(sig.Results().At(0).Type())
			if err := e.extractType(doc.Types, sig.Results().At(0).Type()); err != nil {
				return nil, errors.Wrapf(err, "extracting output type for method %s", method.Name())
			}
		}

		doc.Methods = append(doc.Methods, MethodDoc{
			Name:       method.Name(),
			Comment:    methodDoc,
			File:       methodPos.Filename,
			Line:       methodPos.Line,
			InputType:  inputType,
			OutputType: outputType,
		})
	}

	return doc, nil
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

		// Find the type's file
		var f *ast.File
		if t.Obj().Pkg().Path() == e.pkg.PkgPath {
			for _, file := range e.pkg.Syntax {
				if e.fset.Position(file.Pos()).Filename == pos.Filename {
					f = file
					break
				}
			}
		} else {
			for _, p := range e.pkg.Imports {
				if p.PkgPath == t.Obj().Pkg().Path() {
					for _, file := range p.Syntax {
						if e.fset.Position(file.Pos()).Filename == pos.Filename {
							f = file
							break
						}
					}
				}
			}
		}
		if f == nil {
			fmt.Println("f is nil - ret, looking for '", t.Obj().Pkg().Path(), " - ", t.Obj().Name(), " cur is ", e.pkg.PkgPath)
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

// ExtractType extracts documentation from a type and its fields.
func (e *Extractor) ExtractType(t types.Type) ([]DocComment, error) {
	var comments []DocComment

	switch t := t.(type) {
	case *types.Named:
		pos := e.fset.Position(t.Obj().Pos())
		f, err := parser.ParseFile(e.fset, pos.Filename, nil, parser.ParseComments)
		if err != nil {
			return nil, errors.Wrap(err, "parsing file")
		}

		path, _ := astutil.PathEnclosingInterval(f, t.Obj().Pos(), t.Obj().Pos())
		if len(path) == 0 {
			return nil, nil
		}

		typeSpec, ok := path[0].(*ast.TypeSpec)
		if !ok {
			return nil, nil
		}

		if typeSpec.Doc != nil {
			id := DocID{
				PkgPath:  e.pkg.PkgPath,
				TypeName: t.Obj().Name(),
				FilePath: pos.Filename,
				Line:     pos.Line,
			}
			comments = append(comments, DocComment{
				ID:         id.Hash(),
				Path:       id.PkgPath + "." + id.TypeName,
				Comment:    typeSpec.Doc.Text(),
				SourceFile: pos.Filename,
				Line:       pos.Line,
				Type:       "struct",
				Identifier: id,
			})
		}

		// Extract field documentation if it's a struct
		if st, ok := t.Underlying().(*types.Struct); ok {
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return comments, nil
			}

			for i := 0; i < st.NumFields(); i++ {
				field := st.Field(i)
				if !field.Exported() {
					continue
				}

				astField := structType.Fields.List[i]
				if astField.Doc != nil {
					id := DocID{
						PkgPath:   e.pkg.PkgPath,
						TypeName:  t.Obj().Name(),
						FieldPath: []string{field.Name()},
						FilePath:  pos.Filename,
						Line:      e.fset.Position(astField.Pos()).Line,
					}
					comments = append(comments, DocComment{
						ID:         id.Hash(),
						Path:       id.PkgPath + "." + id.TypeName + "." + field.Name(),
						Comment:    astField.Doc.Text(),
						SourceFile: pos.Filename,
						Line:       e.fset.Position(astField.Pos()).Line,
						Type:       "field",
						Identifier: id,
					})
				}
			}
		}
	}

	return comments, nil
}
