package main

import (
	"flag"
	"go/types"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

const descPkgName = "github.com/utrack/pontoon/sdesc"

func main() {
	dir := flag.String("dir", ".", "directory to parse files from")
	help := flag.Bool("help", false, "print help string and exit")
	outPath := flag.String("out", "", "output file")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	if *outPath == "" {
		*outPath = *dir + "/pontoon.gen.go"
	}

	fout, err := os.Create(*outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

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
	pkgs, err := packages.Load(&pcfg, ".", descPkgName)
	if err != nil {
		log.Fatal(err)
	}

	if len(pkgs) != 2 {
		pkgNames := []string{}
		for _, n := range pkgs {
			pkgNames = append(pkgNames, n.String())
		}
		log.Fatal("more than 2 packages parsed - error? ", len(pkgs), pkgNames)
	}

	var descPkg *packages.Package
	var pkg *packages.Package

	for _, p := range pkgs {
		if p.String() == descPkgName {
			descPkg = p
		}
		pkg = p
	}
	descIface, descMux, err := getDescType(descPkg)
	if err != nil {
		log.Fatal(err)
	}

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

	buf, err := genOpenAPI(svcs, pkg.String())
	if err != nil {
		log.Fatal("when generating OpenAPI 3: ", err)
	}

	res, err := tplGen(tplRequest{
		Content: string(buf),
		PkgPath: pkg.PkgPath,
		PkgName: pkg.Name,
	})
	if err != nil {
		log.Fatal("when executing go code template: ", err)
	}

	_, err = fout.Write(res)
	if err != nil {
		log.Fatal("when writing a file: ", err)
	}

}

func getDescType(pkg *packages.Package) (*types.Interface, *types.Interface, error) {
	decl := pkg.Types.Scope().Lookup("Service")
	t := decl.Type().Underlying().(*types.Interface)

	declMux := pkg.Types.Scope().Lookup("HTTPRouter")
	tMux := declMux.Type().Underlying().(*types.Interface)
	return t, tMux, nil
}
