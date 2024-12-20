package petstore

import (
	"net/http"

	"github.com/ryboe/q"
	"github.com/utrack/pontoon/httpinoapi"
)

// RegisterHandlers registers the pet store handlers with OpenAPI generation
func RegisterHandlers(mux *http.ServeMux, store *Store) error {
	gen := httpinoapi.NewGenerator()

	gen.Operation(http.MethodPost,
		"/pets",
		store.CreatePet,
		httpinoapi.WithInputStruct(CreatePetRequest{}),
		httpinoapi.WithOutputStruct(Pet{}),
	)
	gen.Operation(http.MethodPut,
		"/pets/{id}",
		store.UpdatePet,
		httpinoapi.WithInputStruct(UpdatePetRequest{}),
		httpinoapi.WithOutputStruct(Pet{}),
	)
	gen.Operation(http.MethodGet,
		"/pets",
		store.ListPets,
		httpinoapi.WithInputStruct(ListPetsRequest{}),
		httpinoapi.WithOutputStruct(ListPetsResponse{}),
	)

	// Build OpenAPI document
	doc, err := gen.Build()
	if err != nil {
		return err
	}

	// Set document info
	doc.Info.Title = "Pet Store API"
	doc.Info.Version = "1.0.0"
	doc.Info.Description = "A simple pet store API demonstrating OpenAPI generation"

	// Register handlers with mux
	mux.HandleFunc("/pets", store.CreatePet)
	mux.HandleFunc("/pets/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			store.UpdatePet(w, r)
		} else if r.Method == http.MethodGet {
			store.ListPets(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		q.Q(doc.Paths)
		w.Header().Set("Content-Type", "application/yaml")

		buf, err := doc.Render()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write(buf)
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusFound)
	})
	mux.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		ret := `
		<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Elements in HTML</title>
  
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>

    <elements-api
      apiDescriptionUrl="/openapi.yaml"
      router="hash"
    />

  </body>
</html>
		
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(ret))

	})

	return nil
}
