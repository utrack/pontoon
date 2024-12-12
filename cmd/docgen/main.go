package main

import (
	"errors"
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"path/filepath"

	"github.com/utrack/pontoon/docgen"
	_ "github.com/utrack/pontoon/sdesc"
	"golang.org/x/tools/go/packages"
)

const descPkgName = "github.com/utrack/pontoon/sdesc"

func main() {
	flag.Parse()

	// Configure package loader
	cfg := &packages.Config{
		Mode: packages.NeedImports |
			packages.NeedName |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedModule,
	}

	// Load packages
	pkgs, err := packages.Load(cfg, append(flag.Args(), descPkgName)...)
	if err != nil {
		log.Fatal(err)
	}

	descType, _, err := getDescTypeFromPackages(pkgs)
	if err != nil {
		log.Fatal("failed to load sdesc.Service: " + err.Error())
	}

	// Process each package
	var allFiles []string
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			log.Fatal("Errors when processing Go code: ", p.Errors)
		}
		if p.Name == "sdesc" {
			continue
		}
		if len(p.CompiledGoFiles) == 0 {
			continue
		}
		extractor := docgen.NewExtractor(p)

		var pkgDocs []*docgen.ServiceDoc
		scope := p.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			// New() can return an unexported Handler
			// if !obj.Exported() {
			// 	continue
			// }

			// Check if type implements sdesc.Service
			t, ok := obj.Type().(*types.Named)
			if !ok || !types.Implements(t, descType) {
				continue
			}

			fmt.Printf("%s:%d: found service %s\n",
				p.Fset.Position(obj.Pos()).Filename,
				p.Fset.Position(obj.Pos()).Line,
				obj.Type().String())

			// Extract documentation
			doc, err := extractor.ExtractService(t)
			if err != nil {
				log.Fatalf("%s:%d: failed to extract service %s: %v",
					p.Fset.Position(obj.Pos()).Filename,
					p.Fset.Position(obj.Pos()).Line,
					obj.Type().String(),
					err)
			}

			pkgDocs = append(pkgDocs, doc)
		}

		// Generate a single Go file in the package directory
		// with docs for everything referenced by it
		if len(pkgDocs) > 0 {
			pkgDir := filepath.Dir(p.CompiledGoFiles[0])
			outFile := filepath.Join(pkgDir, "docs_gen.go")

			tomlData, err := docgen.GenerateYAML(pkgDocs, allFiles)
			if err != nil {
				log.Fatal(err)
			}

			goData := docgen.GenerateGoFile(p.Name, tomlData)

			err = os.WriteFile(outFile, goData, 0644)
			if err != nil {
				log.Fatal(err)
			}
			allFiles = append(allFiles, outFile)
		}
	}

}

func getDescTypeFromPackages(pkgs []*packages.Package) (*types.Interface, *types.Interface, error) {

	var descPkg *packages.Package
	for _, p := range pkgs {
		if p.String() == descPkgName {
			descPkg = p
			continue
		}
	}
	if descPkg == nil {
		return nil, nil, errors.New("cannot find package " + descPkgName + " - project doesn't use pontoon?")
	}

	descIface, descMux, err := getDescType(descPkg)
	if err != nil {
		return nil, nil, err
	}
	if descIface == nil {
		return nil, nil, errors.New("couldn't find sdesc.Service definition in " + descPkgName)
	}
	return descIface, descMux, nil
}

func getDescType(pkg *packages.Package) (*types.Interface, *types.Interface, error) {
	decl := pkg.Types.Scope().Lookup("Service")
	if decl == nil {
		return nil, nil, nil
	}
	t := decl.Type().Underlying().(*types.Interface)
	// TODO if struct - old Pontoon!

	declMux := pkg.Types.Scope().Lookup("HTTPRouter")
	tMux := declMux.Type().Underlying().(*types.Interface)
	return t, tMux, nil
}
