# Creating OpenAPI Schemas with libopenapi

This document describes how to create OpenAPI specifications from scratch using the `github.com/pb33f/libopenapi` library.

## Basic Schema Creation

The core building block is the `Schema` object, which is wrapped in a `SchemaProxy` to prevent circular reference issues.

### Creating a Simple Schema

```go
// Use the convenience function to create a SchemaProxy
proxy := base.CreateSchemaProxy(&base.Schema{
    Type: []string{"object"},
    Properties: propMap,
})
```

### Schema with Properties

```go
propMap := orderedmap.New[string, *base.SchemaProxy]()
propMap.Set("nothing", base.CreateSchemaProxy(&base.Schema{
    Type:    []string{"string"},
    Example: &yaml.Node{Value: "nothing"},
}))
```

## Building Complete OpenAPI Specifications

A complete OpenAPI specification consists of:
1. Document metadata (version, info)
2. Paths and operations
3. Responses and media types
4. Components and schemas

### Key Components

- **Document**: The root object containing version and metadata
- **Paths**: Contains all API endpoints
- **Operations**: HTTP methods (GET, POST, etc.) for each path
- **Responses**: Expected responses for operations
- **MediaType**: Content type definitions for requests/responses

## Working with References

References allow reusing schema definitions across the specification.

### Creating References

```go
// Create a reference to a schema
proxy := base.CreateSchemaProxyRef("#/components/schemas/Priority")
```

### Best Practices for References

1. References are supported for Schema definitions
2. Ensure reference paths are correct
3. References should point to components in the same document

## Example Structure

A typical OpenAPI specification structure:

```yaml
openapi: 3.1.0
info:
  title: API Name
  contact:
    name: maintainer
paths:
  /endpoint:
    get:
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResponseType'
components:
  schemas:
    ResponseType:
      type: object
      properties:
        field:
          type: string
```

## Important Notes

1. The library provides high-level models for OpenAPI 3.x
2. Use `orderedmap.New` for maintaining property order
3. All schemas should be wrapped in `SchemaProxy`
4. References are validated at runtime

## Examples

### 1. Basic CRUD API Schema

```go
func CreateCRUDSchema() *v3.Document {
    // User schema
    userProps := orderedmap.New[string, *base.SchemaProxy]()
    userProps.Set("id", base.CreateSchemaProxy(&base.Schema{
        Type:        []string{"integer"},
        Description: "User ID",
    }))
    userProps.Set("name", base.CreateSchemaProxy(&base.Schema{
        Type:        []string{"string"},
        MinLength:   pointer.Int(1),
        MaxLength:   pointer.Int(100),
        Description: "User's full name",
    }))
    userProps.Set("email", base.CreateSchemaProxy(&base.Schema{
        Type:        []string{"string"},
        Format:      "email",
        Description: "User's email address",
    }))

    userSchema := base.CreateSchemaProxy(&base.Schema{
        Type:       []string{"object"},
        Properties: userProps,
        Required:   []string{"name", "email"},
    })

    // Error schema
    errorProps := orderedmap.New[string, *base.SchemaProxy]()
    errorProps.Set("code", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"integer"},
        Example: &yaml.Node{Value: "400"},
    }))
    errorProps.Set("message", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"string"},
        Example: &yaml.Node{Value: "Bad Request"},
    }))

    errorSchema := base.CreateSchemaProxy(&base.Schema{
        Type:       []string{"object"},
        Properties: errorProps,
    })

    // Components
    schemaMap := orderedmap.New[string, *base.SchemaProxy]()
    schemaMap.Set("User", userSchema)
    schemaMap.Set("Error", errorSchema)

    // Response definitions
    successResponse := &v3.Response{
        Description: "Successful operation",
        Content: orderedmap.New[string, *v3.MediaType]().
            Set("application/json", &v3.MediaType{
                Schema: base.CreateSchemaProxyRef("#/components/schemas/User"),
            }).Value(),
    }

    errorResponse := &v3.Response{
        Description: "Error response",
        Content: orderedmap.New[string, *v3.MediaType]().
            Set("application/json", &v3.MediaType{
                Schema: base.CreateSchemaProxyRef("#/components/schemas/Error"),
            }).Value(),
    }

    // Path operations
    pathItemMap := orderedmap.New[string, *v3.PathItem]()
    pathItemMap.Set("/users/{id}", &v3.PathItem{
        Get: &v3.Operation{
            Summary:     "Get user by ID",
            OperationId: "getUserById",
            Parameters: []*v3.Parameter{{
                Name:        "id",
                In:          "path",
                Required:    true,
                Description: "User ID",
                Schema: base.CreateSchemaProxy(&base.Schema{
                    Type: []string{"integer"},
                }),
            }},
            Responses: &v3.Responses{
                Codes: orderedmap.New[string, *v3.Response]().
                    Set("200", successResponse).
                    Set("404", errorResponse).Value(),
            },
        },
    })

    return &v3.Document{
        Version: "3.1.0",
        Info: &base.Info{
            Title:   "User Management API",
            Version: "1.0.0",
        },
        Components: &v3.Components{
            Schemas: schemaMap,
        },
        Paths: &v3.Paths{
            PathItems: pathItemMap,
        },
    }
}
```

### 2. API with Authentication

```go
func CreateAuthenticatedAPI() *v3.Document {
    // Auth header schema
    securitySchemes := orderedmap.New[string, *v3.SecurityScheme]()
    securitySchemes.Set("BearerAuth", &v3.SecurityScheme{
        Type:         "http",
        Scheme:       "bearer",
        BearerFormat: "JWT",
    })

    // Protected endpoint schema
    protectedResponse := &v3.Response{
        Description: "Protected resource",
        Content: orderedmap.New[string, *v3.MediaType]().
            Set("application/json", &v3.MediaType{
                Schema: base.CreateSchemaProxy(&base.Schema{
                    Type: []string{"object"},
                    Properties: orderedmap.New[string, *base.SchemaProxy]().
                        Set("message", base.CreateSchemaProxy(&base.Schema{
                            Type:    []string{"string"},
                            Example: &yaml.Node{Value: "This is a protected resource"},
                        })).Value(),
                }),
            }).Value(),
    }

    pathItemMap := orderedmap.New[string, *v3.PathItem]()
    pathItemMap.Set("/protected", &v3.PathItem{
        Get: &v3.Operation{
            Summary:     "Access protected resource",
            OperationId: "getProtectedResource",
            Security: []map[string][]string{
                {"BearerAuth": {}},
            },
            Responses: &v3.Responses{
                Codes: orderedmap.New[string, *v3.Response]().
                    Set("200", protectedResponse).Value(),
            },
        },
    })

    return &v3.Document{
        Version: "3.1.0",
        Info: &base.Info{
            Title:   "Protected API",
            Version: "1.0.0",
        },
        Components: &v3.Components{
            SecuritySchemes: securitySchemes,
        },
        Paths: &v3.Paths{
            PathItems: pathItemMap,
        },
        Security: []map[string][]string{
            {"BearerAuth": {}},
        },
    }
}
```

### 3. Complex Schema with Arrays and Nested Objects

```go
func CreateComplexSchema() *v3.Document {
    // Address schema
    addressProps := orderedmap.New[string, *base.SchemaProxy]()
    addressProps.Set("street", base.CreateSchemaProxy(&base.Schema{
        Type:     []string{"string"},
        MaxLength: pointer.Int(100),
    }))
    addressProps.Set("city", base.CreateSchemaProxy(&base.Schema{
        Type:     []string{"string"},
        MaxLength: pointer.Int(50),
    }))
    addressProps.Set("country", base.CreateSchemaProxy(&base.Schema{
        Type:     []string{"string"},
        MaxLength: pointer.Int(50),
    }))

    addressSchema := base.CreateSchemaProxy(&base.Schema{
        Type:       []string{"object"},
        Properties: addressProps,
        Required:   []string{"street", "city", "country"},
    })

    // Order item schema
    orderItemProps := orderedmap.New[string, *base.SchemaProxy]()
    orderItemProps.Set("productId", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"string"},
        Pattern: "^[A-Z0-9]{8}$",
    }))
    orderItemProps.Set("quantity", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"integer"},
        Minimum: pointer.Float64(1),
    }))
    orderItemProps.Set("price", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"number"},
        Format:  "float",
        Minimum: pointer.Float64(0),
    }))

    orderItemSchema := base.CreateSchemaProxy(&base.Schema{
        Type:       []string{"object"},
        Properties: orderItemProps,
        Required:   []string{"productId", "quantity", "price"},
    })

    // Complete order schema
    orderProps := orderedmap.New[string, *base.SchemaProxy]()
    orderProps.Set("id", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"string"},
        Pattern: "^ORD-[0-9]{6}$",
    }))
    orderProps.Set("customer", base.CreateSchemaProxy(&base.Schema{
        Type: []string{"object"},
        Properties: orderedmap.New[string, *base.SchemaProxy]().
            Set("name", base.CreateSchemaProxy(&base.Schema{
                Type:     []string{"string"},
                MaxLength: pointer.Int(100),
            })).
            Set("email", base.CreateSchemaProxy(&base.Schema{
                Type:   []string{"string"},
                Format: "email",
            })).
            Set("address", addressSchema).Value(),
        Required: []string{"name", "email", "address"},
    }))
    orderProps.Set("items", base.CreateSchemaProxy(&base.Schema{
        Type:        []string{"array"},
        Items:       orderItemSchema,
        MinItems:    pointer.Int(1),
        Description: "Order items",
    }))
    orderProps.Set("total", base.CreateSchemaProxy(&base.Schema{
        Type:    []string{"number"},
        Format:  "float",
        Minimum: pointer.Float64(0),
    }))

    orderSchema := base.CreateSchemaProxy(&base.Schema{
        Type:       []string{"object"},
        Properties: orderProps,
        Required:   []string{"id", "customer", "items", "total"},
    })

    // Add to components
    schemaMap := orderedmap.New[string, *base.SchemaProxy]()
    schemaMap.Set("Order", orderSchema)
    schemaMap.Set("OrderItem", orderItemSchema)
    schemaMap.Set("Address", addressSchema)

    return &v3.Document{
        Version: "3.1.0",
        Info: &base.Info{
            Title:   "Order Management API",
            Version: "1.0.0",
        },
        Components: &v3.Components{
            Schemas: schemaMap,
        },
    }
}
```

These examples demonstrate:
1. Basic CRUD operations with user management
2. Authentication setup using JWT
3. Complex nested schemas with validation rules
4. Array handling and references
5. Different response types and status codes
6. Parameter validation and formatting

Each example builds on the concepts introduced earlier while showing practical implementations for common API scenarios.

For more details, refer to the [pb33f/libopenapi](https://github.com/pb33f/libopenapi) repository.
