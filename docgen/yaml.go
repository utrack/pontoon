package docgen

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// YAMLDoc represents the YAML documentation format.
type YAMLDoc struct {
	Checksum string                 `yaml:"checksum"`
	Services map[string]YAMLService `yaml:"services"`
	Types    map[string]YAMLType    `yaml:"types"`
}

// YAMLService represents a service in YAML format.
type YAMLService struct {
	ID      string            `yaml:"id"`
	Type    string            `yaml:"type"`
	PkgPath string            `yaml:"pkg_path"`
	Name    string            `yaml:"type_name"`
	File    string            `yaml:"file"`
	Line    int               `yaml:"line"`
	Comment string            `yaml:"comment,omitempty"`
	Methods map[string]Method `yaml:"methods,omitempty"`
}

// Method represents a service method in YAML format.
type Method struct {
	Comment                string `yaml:"comment,omitempty"`
	File                   string `yaml:"file"`
	Line                   int    `yaml:"line"`
	InputType              string `yaml:"input_type,omitempty"`
	OutputType             string `yaml:"output_type,omitempty"`
	ReturnsWellFormedError bool   `yaml:"formed_error,omitempty"`
}

// YAMLType represents a type in YAML format.
type YAMLType struct {
	Package  string      `yaml:"package"`
	Name     string      `yaml:"name"`
	File     string      `yaml:"file"`
	Line     int         `yaml:"line"`
	Comment  string      `yaml:"comment,omitempty"`
	Fields   []YAMLField `yaml:"fields,omitempty"`
	IsStruct bool        `yaml:"is_struct"`
}

// YAMLField represents a field in YAML format.
type YAMLField struct {
	Name       string          `yaml:"name"`
	Type       string          `yaml:"type"`
	Comment    string          `yaml:"comment,omitempty"`
	Tags       string          `yaml:"tags,omitempty"`
	Nullable   bool            `yaml:"nullable,omitempty"`
	IsMap      *YAMLFieldMap   `yaml:"is_map,omitempty"`
	IsArray    *YAMLFieldArray `yaml:"is_array,omitempty"`
	IsEmbedded bool            `yaml:"is_embedded,omitempty"`
}

type YAMLFieldArray struct {
	Type string `yaml:"type"`
}

type YAMLFieldMap struct {
	TypeKey   string `yaml:"key"`
	TypeValue string `yaml:"value"`
}

// GenerateYAML generates YAML documentation from service documentation.
func GenerateYAML(docs []*ServiceDoc, sourceFiles []string) ([]byte, error) {
	// Create YAML document
	yamlDoc := &YAMLDoc{
		Services: make(map[string]YAMLService),
		Types:    make(map[string]YAMLType),
	}

	// First, collect all types from all services
	for _, doc := range docs {
		for typeName, typeDoc := range doc.Types {
			// Only add if not already present or if this one has more information
			existing, exists := yamlDoc.Types[typeName]
			if !exists || (existing.Comment == "" && typeDoc.Comment != "") {
				yamlDoc.Types[typeName] = YAMLType{
					Package:  typeDoc.Package,
					Name:     typeDoc.Name,
					File:     typeDoc.File,
					Line:     typeDoc.Line,
					Comment:  typeDoc.Comment,
					Fields:   make([]YAMLField, len(typeDoc.Fields)),
					IsStruct: typeDoc.IsStruct,
				}
				for i, field := range typeDoc.Fields {
					tf := YAMLField{
						Name:       field.Name,
						Type:       field.Type,
						Comment:    field.Comment,
						Tags:       field.Tags,
						Nullable:   field.Nullable,
						IsEmbedded: field.IsEmbedded,
					}

					if field.IsMap != nil {
						tf.IsMap = &YAMLFieldMap{
							TypeKey:   field.IsMap.TypeKey,
							TypeValue: field.IsMap.TypeValue,
						}
					}
					if field.IsArray != nil {
						tf.IsArray = &YAMLFieldArray{
							Type: field.IsArray.Type,
						}
					}
					yamlDoc.Types[typeName].Fields[i] = tf
				}
			}
		}
	}

	// Then add services
	for _, doc := range docs {
		service := YAMLService{
			ID:      doc.Package + "." + doc.Name,
			Type:    "service",
			PkgPath: doc.Package,
			Name:    doc.Name,
			File:    doc.File,
			Line:    doc.Line,
			Methods: make(map[string]Method),
		}

		// Add comments if present
		if len(doc.Comments) > 0 {
			// Sort comments by line number for stable output
			sort.Slice(doc.Comments, func(i, j int) bool {
				return doc.Comments[i].Line < doc.Comments[j].Line
			})
			service.Comment = doc.Comments[0].Comment
		}

		// Add methods
		for _, m := range doc.Methods {
			service.Methods[m.Name] = Method{
				Comment:                m.Comment,
				File:                   m.File,
				Line:                   m.Line,
				InputType:              m.InputType,
				OutputType:             m.OutputType,
				ReturnsWellFormedError: m.ReturnsWellFormedError,
			}
		}

		yamlDoc.Services[doc.Name] = service
	}

	// Calculate checksum of source files
	h := sha256.New()
	files := make([]string, len(sourceFiles))
	copy(files, sourceFiles)
	sort.Strings(files)
	for _, file := range files {
		h.Write([]byte(file))
	}
	yamlDoc.Checksum = fmt.Sprintf("%x", h.Sum(nil))

	// Marshal to YAML
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(yamlDoc); err != nil {
		return nil, errors.Wrap(err, "encoding YAML")
	}

	return buf.Bytes(), nil
}

func GenerateGoFile(pkgName string, content []byte) []byte {
	ret := `// Code generated by github.com/utrack/pontoon/docgen. DO NOT EDIT.

package ` + pkgName + `

const DocYAML = ` + "`" + string(content) + "`\n"

	return []byte(ret)
}
