package main

import (
	"log"
	"net/http"

	"github.com/utrack/pontoon/examples/petstore"
)

func main() {
	// Create store
	store := petstore.NewStore()

	// Create mux
	mux := http.NewServeMux()

	// Register handlers
	if err := petstore.RegisterHandlers(mux, store); err != nil {
		log.Fatal(err)
	}

	// Start server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
