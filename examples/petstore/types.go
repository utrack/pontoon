package petstore

import "time"

// Pet represents a pet in the store
type Pet struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  Category  `json:"category"`
	Tags      []Tag     `json:"tags"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Category represents a pet category
type Category struct {
	// ID is a unique category ID.
	ID int64 `json:"id"`
	// Name is the category name; non-unique.
	Name string `json:"name"`
}

// Tag represents a pet tag
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Status represents a pet's status in the store
type Status string

const (
	StatusAvailable Status = "available"
	StatusPending   Status = "pending"
	StatusSold      Status = "sold"
)

// CreatePetRequest represents a request to create a new pet
type CreatePetRequest struct {
	// Pet is the pet's details
	Pet CreatePetRequestPet `in:"body=json"`
}

type CreatePetRequestPet struct {
	// Name is the pet's name.
	Name string `json:"name"`
	// Category is this pet's category.
	//
	// A new category is created if it's not found by ID.
	Category Category `json:"category"`
	// Tags are the freeform tags of this pet.
	Tags []Tag `json:"tags"`
	// Status is this pet's status in the store.
	//
	// docgen: enum
	Status Status `json:"status"`
}

// UpdatePetRequest represents a request to update a pet
type UpdatePetRequest struct {
	ID int64 `in:"path=id"`
	// Pet is the new model of this pet.
	Pet CreatePetRequestPet `in:"body=json"`
}

// ListPetsRequest represents a request to list pets
type ListPetsRequest struct {
	// Page is yadda yadda
	Page     int      `in:"query=page"`
	PerPage  int      `in:"query=per_page"`
	Status   Status   `in:"query=status"`
	Category string   `in:"query=category"`
	Tags     []string `in:"query=tags"`
}

// ListPetsResponse represents a response containing a list of pets
type ListPetsResponse struct {
	Items      []Pet `json:"items"`
	TotalItems int64 `json:"total_items"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
