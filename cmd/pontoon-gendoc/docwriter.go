package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/utrack/pontoon/v2/docgen"
)

// docWriter describes a writer that saves the extracted documentation.
// It represents (at least) two different strategies. We either save the files
// alongside the package, or we write them to a single separate package.
type docWriter interface {
	Flush() error
	Write(inFiles []string, pkgName string, typesDocs []*docgen.TypeDoc, funcDocs []*docgen.FunctionDoc) error
}
type writerInPackages struct{}

func (writerInPackages) Flush() error {
	return nil
}

func (writerInPackages) Write(inFiles []string, pName string, typesDocs []*docgen.TypeDoc, funcDocs []*docgen.FunctionDoc) error {

	pkgDir := filepath.Dir(inFiles[0])
	outFile := filepath.Join(pkgDir, "pondoc_gen.go")
	tomlData, err := docgen.GenerateYAML(inFiles, typesDocs, funcDocs)
	if err != nil {
		return errors.Wrap(err, "when generating doc YAML")
	}
	goData := docgen.GenerateGoFile(pName, tomlData)

	err = os.WriteFile(outFile, goData, 0644)
	return errors.Wrap(err, "when writing Go file")
}

type writerSingleDest struct {
	outFile string

	allTypes   []*docgen.TypeDoc
	allFuncs   []*docgen.FunctionDoc
	allInFiles []string
}

func (w *writerSingleDest) Write(inFiles []string, pName string, typesDocs []*docgen.TypeDoc, funcDocs []*docgen.FunctionDoc) error {
	w.allTypes = append(w.allTypes, typesDocs...)
	w.allFuncs = append(w.allFuncs, funcDocs...)
	w.allInFiles = append(w.allInFiles, inFiles...)
	return nil
}
func (w *writerSingleDest) Flush() error {
	absFilePath, err := filepath.Abs(w.outFile)
	if err != nil {
		return errors.Wrapf(err, "when getting absolute path for '%v'", w.outFile)
	}

	absFilePath = filepath.Join(absFilePath, "pondoc_gen.go")
	tomlData, err := docgen.GenerateYAML(w.allInFiles, w.allTypes, w.allFuncs)
	if err != nil {
		return errors.Wrap(err, "when generating doc YAML")
	}

	err = os.MkdirAll(filepath.Dir(absFilePath), 0755)
	if err != nil {
		return errors.Wrapf(err, "when creating directory '%v'", filepath.Dir(absFilePath))
	}

	pName := filepath.Base(filepath.Dir(absFilePath))
	goData := docgen.GenerateGoFile(pName, tomlData)

	err = os.WriteFile(absFilePath, goData, 0644)
	return errors.Wrapf(err, "when writing Go file to '%v'", absFilePath)
}
