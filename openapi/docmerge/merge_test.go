package docmerge

import (
	"fmt"
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/stretchr/testify/require"
	"github.com/utrack/pontoon/docgen"
	ext "github.com/utrack/pontoon/openapi/pontoonext"
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
										e.Set(ext.ExtGoFieldName, &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								p.Set("age", base.CreateSchemaProxy(&base.Schema{
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoFieldName, &yaml.Node{Value: "Age"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "User"})
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
										e.Set(ext.ExtGoFieldName, &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								p.Set("age", base.CreateSchemaProxy(&base.Schema{
									Description: "Age is the user's age in years",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoFieldName, &yaml.Node{Value: "Age"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "User"})
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
										e.Set(ext.ExtGoFieldName, &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "User"})
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
										e.Set(ext.ExtGoFieldName, &yaml.Node{Value: "Name"})
										return e
									}(),
								}))
								return p
							}(),
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "User"})
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
			name: "merge composite type documentation",
			runtime: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							AllOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									SchemaTypeRef: "#/components/schemas/github.com/example/pkg.BaseUser",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "BaseUser"})
										return e
									}(),
								}),
							},
							OneOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "AdminUser"})
										return e
									}(),
								}),
							},
							AnyOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "GuestUser"})
										return e
									}(),
								}),
							},
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "CompositeUser"})
								return e
							}(),
						})
						m.Set("CompositeUser", schema)
						return m
					}(),
				},
			},
			docs: &docgen.YAMLDoc{
				Types: map[string]docgen.YAMLType{
					"github.com/example/pkg.CompositeUser": {
						Package: "github.com/example/pkg",
						Name:    "CompositeUser",
						Comment: "CompositeUser represents a user with multiple roles",
					},
					"github.com/example/pkg.BaseUser": {
						Package: "github.com/example/pkg",
						Name:    "BaseUser",
						Comment: "BaseUser contains common user fields",
					},
					"github.com/example/pkg.AdminUser": {
						Package: "github.com/example/pkg",
						Name:    "AdminUser",
						Comment: "AdminUser represents an administrator",
					},
					"github.com/example/pkg.GuestUser": {
						Package: "github.com/example/pkg",
						Name:    "GuestUser",
						Comment: "GuestUser represents a guest user",
					},
				},
			},
			want: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Description: "CompositeUser represents a user with multiple roles",
							AllOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									Description:   "BaseUser contains common user fields",
									SchemaTypeRef: "#/components/schemas/github.com/example/pkg.BaseUser",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "BaseUser"})
										return e
									}(),
								}),
							},
							OneOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									Description: "AdminUser represents an administrator",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "AdminUser"})
										return e
									}(),
								}),
							},
							AnyOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									Description: "GuestUser represents a guest user",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "GuestUser"})
										return e
									}(),
								}),
							},
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "CompositeUser"})
								return e
							}(),
						})
						m.Set("CompositeUser", schema)
						return m
					}(),
				},
			},
		},
		{
			name: "merge recursive type documentation",
			runtime: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Description: "Initial description",
							AllOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									SchemaTypeRef: "#/components/schemas/github.com/example/pkg.RecursiveType",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "RecursiveType"})
										return e
									}(),
								}),
							},
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "RecursiveType"})
								return e
							}(),
						})
						m.Set("RecursiveType", schema)
						return m
					}(),
				},
			},
			docs: &docgen.YAMLDoc{
				Types: map[string]docgen.YAMLType{
					"github.com/example/pkg.RecursiveType": {
						Package: "github.com/example/pkg",
						Name:    "RecursiveType",
						Comment: "RecursiveType represents a type that references itself",
					},
				},
			},
			want: &v3.Document{
				Components: &v3.Components{
					Schemas: func() *orderedmap.Map[string, *base.SchemaProxy] {
						m := orderedmap.New[string, *base.SchemaProxy]()
						schema := base.CreateSchemaProxy(&base.Schema{
							Description: "RecursiveType represents a type that references itself",
							AllOf: []*base.SchemaProxy{
								base.CreateSchemaProxy(&base.Schema{
									Description:   "RecursiveType represents a type that references itself",
									SchemaTypeRef: "#/components/schemas/github.com/example/pkg.RecursiveType",
									Extensions: func() *orderedmap.Map[string, *yaml.Node] {
										e := orderedmap.New[string, *yaml.Node]()
										e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
										e.Set(ext.ExtGoType, &yaml.Node{Value: "RecursiveType"})
										return e
									}(),
								}),
							},
							Extensions: func() *orderedmap.Map[string, *yaml.Node] {
								e := orderedmap.New[string, *yaml.Node]()
								e.Set(ext.ExtGoPackage, &yaml.Node{Value: "github.com/example/pkg"})
								e.Set(ext.ExtGoType, &yaml.Node{Value: "RecursiveType"})
								return e
							}(),
						})
						m.Set("RecursiveType", schema)
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
			got, err := Merge(tt.runtime, append([]Option{WithDocFile(tt.docs)}, tt.opts...)...)
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
