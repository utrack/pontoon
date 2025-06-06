package petstore

import (
	"net/http"
	"testing"

	"github.com/ryboe/q"
	"github.com/stretchr/testify/require"
	"github.com/utrack/pontoon/v2/httpinoapi"
)

func TestPetStoreOpenAPI(t *testing.T) {
	gen := httpinoapi.NewGenerator()

	// Register handlers
	store := NewStore()
	require.NoError(t,
		gen.Operation(http.MethodPost,
			"/pets",
			store.CreatePet,
			httpinoapi.WithInputStruct(CreatePetRequest{}),
			httpinoapi.WithOutputStruct(Pet{}),
		),
	)
	require.NoError(t,
		gen.Operation(http.MethodPut,
			"/pets/{id}",
			store.UpdatePet,
			httpinoapi.WithInputStruct(UpdatePetRequest{}),
			httpinoapi.WithOutputStruct(Pet{}),
		),
	)
	require.NoError(t,
		gen.Operation(http.MethodGet,
			"/pets",
			store.ListPets,
			httpinoapi.WithInputStruct(ListPetsRequest{}),
			httpinoapi.WithOutputStruct(ListPetsResponse{}),
		),
	)

	// Build OpenAPI document
	doc, err := gen.Build()
	require.NoError(t, err)

	// Verify paths
	require.NotNil(t, doc.Paths)
	require.NotEmpty(t, doc.Paths.PathItems)

	// Verify /pets path
	petsPath, exists := doc.Paths.PathItems.Get("/pets")
	require.True(t, exists)
	require.NotNil(t, petsPath)

	// Verify POST /pets
	require.NotNil(t, petsPath.Post)
	require.NotNil(t, petsPath.Post.RequestBody)
	reqContent, exists := petsPath.Post.RequestBody.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, reqContent.Schema)

	// workaround for OAPI 3.0-style field embeddings
	require.False(t, reqContent.Schema.IsReference())
	require.True(t, len(reqContent.Schema.Schema().AllOf) == 1)
	require.True(t, reqContent.Schema.Schema().AllOf[0].IsReference())

	// Verify PUT /pets/{id}
	petsIDPath, exists := doc.Paths.PathItems.Get("/pets/{id}")
	require.True(t, exists)
	require.NotNil(t, petsIDPath)
	require.NotNil(t, petsIDPath.Put)
	require.NotEmpty(t, petsIDPath.Put.Parameters)

	// Verify GET /pets
	require.NotNil(t, petsPath.Get)
	q.Q(petsPath.Get)
	require.NotEmpty(t, petsPath.Get.Parameters)
	require.NotNil(t, petsPath.Get.Responses)
	resp, exists := petsPath.Get.Responses.Codes.Get("200")
	require.True(t, exists)
	require.NotNil(t, resp)
	respContent, exists := resp.Content.Get("application/json")
	require.True(t, exists)
	require.NotNil(t, respContent.Schema)
	require.True(t, respContent.Schema.IsReference())

	// Verify components
	require.NotNil(t, doc.Components)
	require.NotEmpty(t, doc.Components.Schemas)

	// Verify Pet schema
	q.Q(doc.Components.Schemas)
	petSchema, exists := doc.Components.Schemas.Get("github.com_utrack_pontoon_examples_petstore.Pet")
	require.True(t, exists)
	require.NotNil(t, petSchema)
	require.Equal(t, []string{"object"}, petSchema.Schema().Type)
	require.NotNil(t, petSchema.Schema().Properties)

	// Verify Category schema
	categorySchema, exists := doc.Components.Schemas.Get("github.com_utrack_pontoon_examples_petstore.Category")
	require.True(t, exists)
	require.NotNil(t, categorySchema)
	require.Equal(t, []string{"object"}, categorySchema.Schema().Type)
	require.NotNil(t, categorySchema.Schema().Properties)

	// Verify Tag schema
	tagSchema, exists := doc.Components.Schemas.Get("github.com_utrack_pontoon_examples_petstore.Tag")
	require.True(t, exists)
	require.NotNil(t, tagSchema)
	require.Equal(t, []string{"object"}, tagSchema.Schema().Type)
	require.NotNil(t, tagSchema.Schema().Properties)
}
