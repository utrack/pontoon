package petstore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ggicci/httpin"
)

// Store is a simple in-memory pet store
type Store struct {
	pets map[int64]Pet
	seq  int64
}

// NewStore creates a new pet store
func NewStore() *Store {
	return &Store{
		pets: make(map[int64]Pet),
	}
}

// CreatePet creates a new pet
func (s *Store) CreatePet(w http.ResponseWriter, r *http.Request) {
	var req CreatePetRequest

	err := httpin.DecodeTo(r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	s.seq++
	pet := Pet{
		ID:        s.seq,
		Name:      req.Pet.Name,
		Category:  req.Pet.Category,
		Tags:      req.Pet.Tags,
		Status:    req.Pet.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.pets[pet.ID] = pet

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pet)
}

// UpdatePet updates an existing pet
func (s *Store) UpdatePet(w http.ResponseWriter, r *http.Request) {
	var req UpdatePetRequest
	err := httpin.DecodeTo(r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	pet, exists := s.pets[req.ID]
	if !exists {
		writeError(w, http.StatusNotFound, "Pet not found", fmt.Sprintf("Pet with ID %d not found", req.ID))
		return
	}

	pet.Name = req.Pet.Name
	pet.Category = req.Pet.Category
	pet.Tags = req.Pet.Tags
	pet.Status = req.Pet.Status
	pet.UpdatedAt = time.Now()
	s.pets[pet.ID] = pet

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pet)
}

// ListPets lists all pets matching the given criteria
func (s *Store) ListPets(w http.ResponseWriter, r *http.Request) {
	var req ListPetsRequest

	err := httpin.DecodeTo(r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Apply defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 10
	}

	// Filter pets
	var filtered []Pet
	for _, pet := range s.pets {
		if req.Status != "" && pet.Status != req.Status {
			continue
		}
		if req.Category != "" && pet.Category.Name != req.Category {
			continue
		}
		if len(req.Tags) > 0 {
			hasTag := false
			for _, tag := range pet.Tags {
				for _, reqTag := range req.Tags {
					if tag.Name == reqTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		filtered = append(filtered, pet)
	}

	// Calculate pagination
	start := (req.Page - 1) * req.PerPage
	end := start + req.PerPage
	if end > len(filtered) {
		end = len(filtered)
	}
	if start > len(filtered) {
		start = len(filtered)
	}

	resp := ListPetsResponse{
		Items:      filtered[start:end],
		TotalItems: int64(len(filtered)),
		Page:       req.Page,
		PerPage:    req.PerPage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, code int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
