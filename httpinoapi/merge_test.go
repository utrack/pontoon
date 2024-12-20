package httpinoapi

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestRequest struct {
	Name  string `in:"query=str"`
	Bytes []byte `in:"body=json"`
}

type TestRequestWithBody struct {
	Name string `in:"query=str"`

	RequestBody TestRequest `in:"body=json"`
}

type TestResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Complex types for testing schema generation
type QueryParams struct {
	Page    int      `in:"query=page"`
	PerPage int      `in:"query=per_page"`
	SortBy  string   `in:"query=sort"`
	Filters []string `in:"query=filter"`
}

type RequestBody struct {
	Data map[string]string `in:"body=json"`
}

type ComplexRequest struct {
	QueryParams
	RequestBody
}

type NestedType struct {
	ID        int               `json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata"`
}

type ResponseData struct {
	Items      []NestedType          `json:"items"`
	Pagination map[string]int        `json:"pagination"`
	Extra      map[string]NestedType `json:"extra"`
}

type ComplexResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    ResponseData `json:"data"`
}

type ErrorDetails struct {
	Field   string `json:"field"`
	Problem string `json:"problem"`
}

type ErrorResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Details ErrorDetails `json:"details"`
}

func TestGenerator_Build(t *testing.T) {
	g := NewGenerator()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// Register a GET endpoint
	err := g.Operation(http.MethodGet, "/users/:id", handler,
		WithTags("users"),
		WithDescription("Get user by ID"),
		WithOutputStruct(TestResponse{}))
	require.NoError(t, err)

	// Register a POST endpoint
	err = g.Operation(http.MethodPost, "/users", handler,
		WithTags("users"),
		WithDescription("Create user"),
		WithInputStruct(TestRequest{}),
		WithOutputStruct(TestResponse{}))
	require.NoError(t, err)

	// Build OpenAPI document
	doc, err := g.Build()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Verify OpenAPI version
	require.Equal(t, "3.1.0", doc.Version)

	// Verify paths
	getPath, exists := doc.Paths.PathItems.Get("/users/:id")
	require.True(t, exists)
	require.NotNil(t, getPath)

	postPath, exists := doc.Paths.PathItems.Get("/users")
	require.True(t, exists)
	require.NotNil(t, postPath)

	// Verify GET operation
	require.NotNil(t, getPath.Get)
	require.Equal(t, []string{"users"}, getPath.Get.Tags)
	require.Equal(t, "Get user by ID", getPath.Get.Description)
	require.NotNil(t, getPath.Get.Responses, "GET Responses is nil")
	require.NotNil(t, getPath.Get.Responses.Codes, "GET Responses.Codes is nil")
	resp, exists := getPath.Get.Responses.Codes.Get("200")
	fmt.Printf("GET Response exists: %v\n", exists)
	if exists {
		fmt.Printf("GET Response: %+v\n", resp)
	}
	require.True(t, exists)
	require.NotNil(t, resp)

	// Verify POST operation
	require.NotNil(t, postPath.Post)
	require.Equal(t, []string{"users"}, postPath.Post.Tags)
	require.Equal(t, "Create user", postPath.Post.Description)
	require.NotNil(t, postPath.Post.RequestBody)
	require.NotNil(t, postPath.Post.Responses, "POST Responses is nil")
	require.NotNil(t, postPath.Post.Responses.Codes, "POST Responses.Codes is nil")
	resp, exists = postPath.Post.Responses.Codes.Get("200")
	fmt.Printf("POST Response exists: %v\n", exists)
	if exists {
		fmt.Printf("POST Response: %+v\n", resp)
	}
	require.True(t, exists)
	require.NotNil(t, resp)
}

func TestGenerator_Build_ComponentReferences(t *testing.T) {
	g := NewGenerator()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// Register a POST endpoint with request/response bodies
	err := g.Operation(http.MethodPost, "/users", handler,
		WithInputStruct(TestRequestWithBody{}),
		WithOutputStruct(TestResponse{}))
	require.NoError(t, err)

	// Build OpenAPI document
	doc, err := g.Build()
	require.NoError(t, err)

	// Get POST operation
	postPath, exists := doc.Paths.PathItems.Get("/users")
	require.True(t, exists)
	require.NotNil(t, postPath.Post)

	// Verify request body is a reference
	require.NotNil(t, postPath.Post.RequestBody)
	reqContent, exists := postPath.Post.RequestBody.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, reqContent.Schema)

	require.NotEmpty(t, reqContent.Schema.Schema().SchemaTypeRef, "Request body schema should be a reference")
	require.True(t, strings.HasPrefix(reqContent.Schema.Schema().SchemaTypeRef, "#/components/schemas/"),
		"Request body reference should point to Components")

	// Verify response body is a reference
	resp, exists := postPath.Post.Responses.Codes.Get("200")
	require.True(t, exists)
	respContent, exists := resp.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, respContent.Schema)
	require.True(t, respContent.Schema.IsReference())
	require.NotEmpty(t, respContent.Schema.GetReference(), "Response body schema should be a reference")
	require.True(t, strings.HasPrefix(respContent.Schema.GetReference(), "#/components/schemas/"),
		"Response body reference should point to Components")
}

func TestGenerator_Build_ComplexTypes(t *testing.T) {
	g := NewGenerator()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// Register endpoints with complex types
	err := g.Operation(http.MethodGet, "/items", handler,
		WithDescription("List items with complex response"),
		WithOutputStruct(ComplexResponse{}))
	require.NoError(t, err)

	err = g.Operation(http.MethodPost, "/items", handler,
		WithDescription("Create item with complex request"),
		WithInputStruct(ComplexRequest{}),
		WithOutputStruct(ComplexResponse{}))
	require.NoError(t, err)

	// Build OpenAPI document
	doc, err := g.Build()
	require.NoError(t, err)

	// Verify GET operation
	getPath, exists := doc.Paths.PathItems.Get("/items")
	require.True(t, exists)
	require.NotNil(t, getPath.Get)

	// Verify response schema is a reference
	resp, exists := getPath.Get.Responses.Codes.Get("200")
	require.True(t, exists)
	respContent, exists := resp.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, respContent.Schema)
	require.True(t, respContent.Schema.IsReference())
	require.NotEmpty(t, respContent.Schema.GetReference())

	// Verify POST operation
	require.NotNil(t, getPath.Post)
	require.NotNil(t, getPath.Post.RequestBody)

	// Verify request body schema is a reference
	reqContent, exists := getPath.Post.RequestBody.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, reqContent.Schema)
	// this is a map[string]string, which is done via additionalProperties
	// the string scalar is not referenced
	require.NotNil(t, reqContent.Schema.Schema().AdditionalProperties)
	require.NotNil(t, reqContent.Schema.Schema().AdditionalProperties.A.Schema())
	require.Equal(t, []string{"string"}, reqContent.Schema.Schema().AdditionalProperties.A.Schema().Type)
}

func TestGenerator_Build_ErrorResponses(t *testing.T) {
	g := NewGenerator()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// Register endpoint with error response
	err := g.Operation(http.MethodPost, "/items", handler,
		WithDescription("Create item with error handling"),
		WithInputStruct(TestRequest{}),
		WithOutputStruct(TestResponse{}))
	require.NoError(t, err)

	// Build OpenAPI document
	doc, err := g.Build()
	require.NoError(t, err)

	// Get POST operation
	postPath, exists := doc.Paths.PathItems.Get("/items")
	require.True(t, exists)
	require.NotNil(t, postPath.Post)

	// Verify success response
	resp, exists := postPath.Post.Responses.Codes.Get("200")
	require.True(t, exists)
	require.NotNil(t, resp)
	require.Equal(t, "Successful response", resp.Description)

	// Verify response schema is a reference
	respContent, exists := resp.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, respContent.Schema)
	require.True(t, respContent.Schema.IsReference())
	require.NotEmpty(t, respContent.Schema.GetReference())
}

func TestGenerator_Build_EdgeCases(t *testing.T) {
	g := NewGenerator()
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// Test empty struct
	err := g.Operation(http.MethodGet, "/empty", handler,
		WithDescription("Empty response struct"),
		WithOutputStruct(struct{}{}))
	require.NoError(t, err)

	// Test nil input/output
	err = g.Operation(http.MethodGet, "/nil", handler,
		WithDescription("Nil input/output"))
	require.NoError(t, err)

	// Test duplicate type registration
	err = g.Operation(http.MethodPost, "/dup1", handler,
		WithInputStruct(TestRequest{}),
		WithOutputStruct(TestResponse{}))
	require.NoError(t, err)

	err = g.Operation(http.MethodPost, "/dup2", handler,
		WithInputStruct(TestRequest{}),
		WithOutputStruct(TestResponse{}))
	require.NoError(t, err)

	// Build OpenAPI document
	doc, err := g.Build()
	require.NoError(t, err)

	// Verify empty struct endpoint
	emptyPath, exists := doc.Paths.PathItems.Get("/empty")
	require.True(t, exists)
	require.NotNil(t, emptyPath.Get)

	// Verify nil endpoint
	nilPath, exists := doc.Paths.PathItems.Get("/nil")
	require.True(t, exists)
	require.NotNil(t, nilPath.Get)

	// Verify duplicate endpoints share the same schema references
	dup1Path, exists := doc.Paths.PathItems.Get("/dup1")
	require.True(t, exists)
	dup2Path, exists := doc.Paths.PathItems.Get("/dup2")
	require.True(t, exists)

	// Get schema refs
	dup1Content, _ := dup1Path.Post.RequestBody.Content.Get("application/json")
	dup2Content, _ := dup2Path.Post.RequestBody.Content.Get("application/json")
	require.Equal(t,
		dup1Content.Schema.Schema().SchemaTypeRef,
		dup2Content.Schema.Schema().SchemaTypeRef,
		"Duplicate types should share the same schema reference")
}
