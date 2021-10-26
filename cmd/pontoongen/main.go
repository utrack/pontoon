package main

import (
	"flag"
	"go/types"
	"log"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
)

const descPkgName = "github.com/utrack/pontoon/sdesc"

func main() {
	dir := flag.String("dir", ".", "directory to parse files from")
	help := flag.Bool("help", false, "print help string and exit")
	recursive := flag.Bool("recursive", false, "generate defs for all child modules recursively")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	pcfg := packages.Config{
		Mode: packages.NeedTypesInfo |
			packages.NeedSyntax |
			packages.NeedName |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedModule |
			packages.NeedExportsFile,
		Dir: *dir,
	}
	parsePath := "."
	if *recursive {
		parsePath = "./..."
	}
	srcPkgs, err := packages.Load(&pcfg, parsePath, descPkgName)
	if err != nil {
		log.Fatal(err)
	}

	if len(srcPkgs) < 2 {
		pkgNames := []string{}
		for _, n := range srcPkgs {
			pkgNames = append(pkgNames, n.String())
		}
		log.Fatal("less than 2 packages parsed - error? ", len(srcPkgs), pkgNames)
	}

	var descPkg *packages.Package
	pkgs := []*packages.Package{}

	for i, p := range srcPkgs {
		if p.String() == descPkgName {
			descPkg = p
			continue
		}
		pkgs = append(pkgs, srcPkgs[i])
	}

	if descPkg == nil {
		log.Fatal("cannot find package " + descPkgName)
	}
	descIface, descMux, err := getDescType(descPkg)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range pkgs {
		bu := builder{pkg: pkg, muxType: descMux}

		scope := pkg.Types.Scope()
		svcs := []serviceDesc{}
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			namedT, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}

			_, ok = namedT.Underlying().(*types.Struct)
			if !ok {
				continue
			}
			ptr := types.NewPointer(namedT)
			imp := types.Implements(ptr.Underlying(), descIface)
			if !imp {
				continue
			}

			ms := types.NewMethodSet(namedT)
			svc, err := bu.Service(ms, namedT, pkg.Fset)
			if err != nil {
				log.Fatal(err)
			}
			svcs = append(svcs, *svc)
		}

		for _, svc := range svcs {
			buf, err := genOpenAPI(svcs, pkg.String())
			if err != nil {
				log.Fatal("when generating OpenAPI 3: ", err)
			}

			if !filepath.IsAbs(svc.filename) {
				panic(svc.filename + "<- path is not absolute")
			}

			dir := filepath.Dir(svc.filename)
			path := filepath.Join(dir, strcase.ToSnake(svc.serviceStructName)+".pontoon.go")

			res, err := tplGen(tplRequest{
				Content:           string(buf),
				PkgPath:           pkg.PkgPath,
				PkgName:           svc.pkg,
				HandlerStructName: svc.serviceStructName,
			})
			if err != nil {
				log.Fatal("when executing go code template: ", err)
			}
			fout, err := os.Create(path)
			if err != nil {
				log.Fatal(err)
			}

			defer fout.Close()
			_, err = fout.Write(res)
			if err != nil {
				log.Fatal("when writing a file: ", err)
			}
			fout.Close()
		}
	}

}

func getDescType(pkg *packages.Package) (*types.Interface, *types.Interface, error) {
	decl := pkg.Types.Scope().Lookup("Service")
	t := decl.Type().Underlying().(*types.Interface)

	declMux := pkg.Types.Scope().Lookup("HTTPRouter")
	tMux := declMux.Type().Underlying().(*types.Interface)
	return t, tMux, nil
}
