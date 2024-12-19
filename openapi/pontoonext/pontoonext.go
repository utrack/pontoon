package pontoonext

import (
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

// GoTypeInfo encapsulates Go type information from OpenAPI extensions
type GoTypeInfo struct {
	PackagePath string // Go package import path
	TypeName    string // Go type name
}

func (g GoTypeInfo) String() string {
	if g.PackagePath == "" {
		return g.TypeName
	}

	return g.PackagePath + "." + g.TypeName
}

// ExtensionKeys contains known extension keys
const (
	ExtGoPackage   = "x-pontoon-go-package"
	ExtGoType      = "x-pontoon-go-type"
	ExtGoFieldName = "x-pontoon-field-go-name"
)

func (g GoTypeInfo) SetTo(extensions *orderedmap.Map[string, *yaml.Node]) {
	if g.PackagePath != "" {
		extensions.Set(ExtGoPackage, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: g.PackagePath,
		})
	}
	if g.TypeName != "" {
		extensions.Set(ExtGoType, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: g.TypeName,
		})
	}
}

// GetGoTypeInfo extracts Go type information from schema extensions
func GetGoTypeInfo(extensions *orderedmap.Map[string, *yaml.Node]) (*GoTypeInfo, bool) {
	if extensions == nil {
		return nil, false
	}

	pkgPath, ok := extensions.Get(ExtGoPackage)
	if !ok {
		return nil, false
	}
	typeName, ok := extensions.Get(ExtGoType)
	if !ok {
		return nil, false
	}

	return &GoTypeInfo{
		PackagePath: pkgPath.Value,
		TypeName:    typeName.Value,
	}, true
}

// GetGoFieldName extracts Go field name from schema extensions
func GetGoFieldName(extensions *orderedmap.Map[string, *yaml.Node]) (string, bool) {
	if extensions == nil {
		return "", false
	}
	fieldName, ok := extensions.Get(ExtGoFieldName)
	if !ok {
		return "", false
	}
	return fieldName.Value, true
}
