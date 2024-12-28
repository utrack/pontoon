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

var (
	allFlag = flag.Bool("all", false, "Generate documentation for all types in package, not just Service implementations")
)

func buildPkgMap(m map[string]*packages.Package, p *packages.Package) {
	if p == nil {
		return
	}
	if _, ok := m[p.PkgPath]; ok {
		return
	}

	m[p.PkgPath] = p
	for _, imp := range p.Imports {
		buildPkgMap(m, imp)
	}
}

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

	// Traverse the imports map and build a
	// complete pkgMap
	allPkgs := make(map[string]*packages.Package)
	for _, p := range pkgs {
		buildPkgMap(allPkgs, p)
	}

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
		extractor := docgen.NewExtractor(p, allPkgs)

		var typesDocs []*docgen.TypeDoc
		var funcDocs []*docgen.FunctionDoc

		scope := p.Types.Scope()

		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			switch t := obj.Type().(type) {
			case *types.Named:

				var err error

				if !*allFlag && !types.Implements(t, descType) {
					continue
				}

				fmt.Printf("%s:%d: found type %s\n",
					p.Fset.Position(obj.Pos()).Filename,
					p.Fset.Position(obj.Pos()).Line,
					obj.Type().String())

				// Extract type documentation
				types, err := extractor.ExtractType(t)
				if err != nil {
					fmt.Printf("Warning: %s:%d: failed to extract comments for %s: %v\n",
						p.Fset.Position(obj.Pos()).Filename,
						p.Fset.Position(obj.Pos()).Line,
						obj.Type().String(),
						err)
					continue
				}

				for k := range types {
					v := types[k]
					typesDocs = append(typesDocs, &v)
				}
			case *types.Signature:
				if !*allFlag {
					continue
				}

				fmt.Printf("%s:%d: found function %s\n",
					p.Fset.Position(obj.Pos()).Filename,
					p.Fset.Position(obj.Pos()).Line,
					obj.Name())
				f, err := extractor.ExtractFunction(t, obj)
				if err != nil {
					fmt.Printf("Warning: %s:%d: failed to extract comments for %s: %v\n",
						p.Fset.Position(obj.Pos()).Filename,
						p.Fset.Position(obj.Pos()).Line,
						obj.Type().String(),
						err)
					continue
				}
				funcDocs = append(funcDocs, f)
			default:
				// TODO log if debug on
			}
		}

		// Generate a single Go file in the package directory
		// with docs for everything referenced by it
		if len(typesDocs) > 0 {
			pkgDir := filepath.Dir(p.CompiledGoFiles[0])
			outFile := filepath.Join(pkgDir, "docs_gen.go")

			tomlData, err := docgen.GenerateYAML(p.CompiledGoFiles, typesDocs,funcDocs)
			if err != nil {
				log.Fatal(err)
			}

			goData := docgen.GenerateGoFile(p.Name, tomlData)

			err = os.WriteFile(outFile, goData, 0644)
			if err != nil {
				log.Fatal(err)
			}
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
