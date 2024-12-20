# Pet Store Example

This example demonstrates how to use the OpenAPI generation features of Pontoon to create a simple pet store API.

## Features Demonstrated

1. Request/Response Types:
   - Request types with `in` tags for query parameters, path parameters, and JSON body
   - Response types with JSON tags
   - Error response types
   - Embedded types (Category, Tag)
   - Array types (Tags)
   - Enum types (Status)

2. OpenAPI Generation:
   - Handler registration with HTTP methods and paths
   - Request body schema generation
   - Query parameter schema generation
   - Path parameter schema generation
   - Response schema generation
   - Component schema generation
   - Schema references

3. HTTP Handler Implementation:
   - Request parsing
   - Response writing
   - Error handling
   - Query parameter parsing
   - Path parameter parsing
   - JSON request/response handling

## Running the Example

```bash
# Run tests
go test -v ./examples/petstore/...

# Start the server
go run ./examples/petstore/cmd/server/main.go
```

## API Endpoints

1. `POST /pets`
   - Creates a new pet
   - Request: JSON body with pet details
   - Response: Created pet with ID

2. `PUT /pets/{id}`
   - Updates an existing pet
   - Request: Path parameter for ID, JSON body with pet details
   - Response: Updated pet

3. `GET /pets`
   - Lists pets with filtering and pagination
   - Query parameters:
     - `page`: Page number (default: 1)
     - `per_page`: Items per page (default: 10)
     - `status`: Filter by status (available, pending, sold)
     - `category`: Filter by category name
     - `tags`: Filter by tag names (comma-separated)
   - Response: List of pets with pagination info
