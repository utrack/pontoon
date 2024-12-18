package docmerge

import (
	"fmt"
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/stretchr/testify/require"
	"github.com/utrack/pontoon/docgen"
	"gopkg.in/yaml.v3"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name    string
		runtime *v3.Document
		docs    *docgen.YAMLDoc
		opts    []Option
		want    *v3.Document
		wantErr bool
	}{
		{
			name: "merge type documentation",
			runtime: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Properties: func() *orderedmap.Map[string, *base.SchemaProxy] {
								p := orderedmap.New[string, *base.SchemaProxy]()
								p.Set("name", base.CreateSchemaProxy(&base.Schema{
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set("x-pontoon-field-go-name", &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								p.Set("age", base.CreateSchemaProxy(&base.Schema{
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set("x-pontoon-field-go-name", &yaml.Node{Value: "Age"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set("x-pontoon-go-package", &yaml.Node{Value: "github.com/example/pkg"})
								e.Set("x-pontoon-go-type", &yaml.Node{Value: "User"})
								return e
							}(),
						})
						m.Set("User", schema)
						return m
					}(),
				},
			},
			docs: &docgen.YAMLDoc{
				Types: map[string]docgen.YAMLType{
					"github.com/example/pkg.User": {
						Package: "github.com/example/pkg",
						Name:    "User",
						Comment: "User represents a user in the system",
						Fields: []docgen.YAMLField{
							{
								Name:    "Name",
								Comment: "Name is the user's full name",
							},
							{
								Name:    "Age",
								Comment: "Age is the user's age in years",
							},
						},
					},
				},
			},
			want: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Description: "User represents a user in the system",
							Properties: func() *orderedmap.Map[string, *base.SchemaProxy] {
								p := orderedmap.New[string, *base.SchemaProxy]()
								p.Set("name", base.CreateSchemaProxy(&base.Schema{
									Description: "Name is the user's full name",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set("x-pontoon-field-go-name", &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								p.Set("age", base.CreateSchemaProxy(&base.Schema{
									Description: "Age is the user's age in years",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set("x-pontoon-field-go-name", &yaml.Node{Value: "Age"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set("x-pontoon-go-package", &yaml.Node{Value: "github.com/example/pkg"})
								e.Set("x-pontoon-go-type", &yaml.Node{Value: "User"})
								return e
							}(),
						})
						m.Set("User", schema)
						return m
					}(),
				},
			},
		},
		{
			name: "preserve existing documentation",
			runtime: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Description: "Existing description",
							Properties: func() *orderedmap.Map[string, *base.SchemaProxy] {
								p := orderedmap.New[string, *base.SchemaProxy]()
								p.Set("name", base.CreateSchemaProxy(&base.Schema{
									Description: "Existing field description",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set("x-pontoon-field-go-name", &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set("x-pontoon-go-package", &yaml.Node{Value: "github.com/example/pkg"})
								e.Set("x-pontoon-go-type", &yaml.Node{Value: "User"})
								return e
							}(),
						})
						m.Set("User", schema)
						return m
					}(),
				},
			},
			docs: &docgen.YAMLDoc{
				Types: map[string]docgen.YAMLType{
					"github.com/example/pkg.User": {
						Package: "github.com/example/pkg",
						Name:    "User",
						Comment: "New description",
						Fields: []docgen.YAMLField{
							{
								Name:    "Name",
								Comment: "New field description",
							},
						},
					},
				},
			},
			opts: []Option{WithPreserveExisting(true)},
			want: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Description: "Existing description",
							Properties: func() *orderedmap.Map[string, *base.SchemaProxy] {
								p := orderedmap.New[string, *base.SchemaProxy]()
								p.Set("name", base.CreateSchemaProxy(&base.Schema{
									Description: "Existing field description",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set("x-pontoon-field-go-name", &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set("x-pontoon-go-package", &yaml.Node{Value: "github.com/example/pkg"})
								e.Set("x-pontoon-go-type", &yaml.Node{Value: "User"})
								return e
							}(),
						})
						m.Set("User", schema)
						return m
					}(),
				},
			},
		},
		{
			name:    "empty documents",
			runtime: nil,
			docs:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Merge(tt.runtime, tt.docs, tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			gotYaml, err := got.MarshalYAML()
			require.NoError(t, err)
			gotBytes, err := yaml.Marshal(gotYaml)
			require.NoError(t, err)
			fmt.Println(string(gotBytes))


			wantYaml, err := tt.want.MarshalYAML()
			require.NoError(t, err)
			wantBytes, err := yaml.Marshal(wantYaml)
			require.NoError(t, err)

			fmt.Println(string(wantBytes))

			require.EqualValues(t, string(wantBytes), string(gotBytes))
		})
	}
}
