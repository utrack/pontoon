package httpinoapi

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/utrack/pontoon/openapi/docmerge"
	"github.com/utrack/pontoon/openapi/httpinmeditate"
)

// buildOperation creates an OpenAPI operation from a handler.
func (g *Generator) buildOperation(h *handlerInfo) (*v3.Operation, error) {
	op := &v3.Operation{
		Tags:        h.options.tags,
		Description: h.options.description,
	}

	oapigen := httpinmeditate.NewGenerator()

	// Generate request parameters if input type is specified
	if h.options.inputType != nil {
		reqSchema, reqParams, err := oapigen.GenerateOperationRequestParams(h.options.inputType)
		if err != nil {
			return nil, fmt.Errorf("generate request params: %w", err)
		}

		if reqSchema != nil {
			if !reqSchema.IsReference() && reqSchema.Schema().SchemaTypeRef != "" {
				reqSchema = base.CreateSchemaProxyRef(reqSchema.Schema().SchemaTypeRef)
			}
			op.RequestBody = &v3.RequestBody{
				Content: orderedmap.New[string, *v3.MediaType](),
			}
			op.RequestBody.Content.Set("application/json", &v3.MediaType{
				Schema: reqSchema,
			})
		}

		if len(reqParams) > 0 {
			op.Parameters = reqParams
		}
	}

	// Generate response if output type is specified
	if h.options.outputType != nil {
		respSchema, err := oapigen.JSONSchemaRef(h.options.outputType)

		if err != nil {
			return nil, fmt.Errorf("generate response schema: %w", err)
		}

		if !respSchema.IsReference() && respSchema.Schema().SchemaTypeRef != "" {
			respSchema = base.CreateSchemaProxyRef(respSchema.Schema().SchemaTypeRef)
		}

		op.Responses = &v3.Responses{
			Codes: orderedmap.New[string, *v3.Response](),
		}
		resp := &v3.Response{
			Description: "Successful response",
			Content:     orderedmap.New[string, *v3.MediaType](),
		}
		resp.Content.Set("application/json", &v3.MediaType{
			Schema: respSchema,
		})
		op.Responses.Codes.Set("200", resp)
	}

	if g.components == nil || g.components.Schemas == nil {
		g.components = oapigen.Components()
		g.components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}
	// TODO compare if components are the same on conflict
	for k, v := range oapigen.Components().Schemas.FromNewest() {
		fmt.Println("extract from oapi - schema ", k)
		g.components.Schemas.Set(k, v)
	}
	// TODO there might be more than just Schemas

	return op, nil
}

// Build returns a complete OpenAPI 3.1 document with all registered handlers.
func (g *Generator) Build() (*v3.Document, error) {
	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title:   "API Documentation",
			Version: "1.0.0",
		},
		Paths: &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		},
	}

	// Build operations for each handler
	for _, h := range g.handlers {
		op, err := g.buildOperation(h)
		if err != nil {
			return nil, fmt.Errorf("build operation for %s %s: %w", h.verb, h.path, err)
		}

		// Create path item if it doesn't exist
		pathItem, exists := doc.Paths.PathItems.Get(h.path)
		if !exists {
			pathItem = &v3.PathItem{}
			doc.Paths.PathItems.Set(h.path, pathItem)
		}

		// Set operation based on HTTP verb
		switch h.verb {
		case "GET":
			pathItem.Get = op
		case "POST":
			pathItem.Post = op
		case "PUT":
			pathItem.Put = op
		case "DELETE":
			pathItem.Delete = op
		case "PATCH":
			pathItem.Patch = op
		case "HEAD":
			pathItem.Head = op
		case "OPTIONS":
			pathItem.Options = op
		case "TRACE":
			pathItem.Trace = op
		default:
			return nil, fmt.Errorf("unsupported HTTP verb: %s", h.verb)
		}
	}
	doc.Components = g.components

	// Merge with existing documentation if any
	mergedDoc, err := docmerge.Merge(doc)
	if err != nil {
		return nil, fmt.Errorf("merge documentation: %w", err)
	}

	// buf, err := doc.Render()
	// if err != nil {
	// 	return nil, fmt.Errorf("render document: %w", err)
	// }
	// parsedDoc, err := libopenapi.NewDocument(buf)
	// if err != nil {
	// 	//return nil, errors.Wrap(err, "failed to back-parse the document")
	// }
	// validator, errs := vpkg.NewValidator(parsedDoc)
	// if len(errs) > 0 {
	// 	return nil, errors.Errorf("failed to create validator: %v", errs)
	// }
	// valid, validationErrs := validator.ValidateDocument()
	// if !valid {
	// 	return nil, fmt.Errorf("validation errors: %v", validationErrs)
	// }

	return mergedDoc, nil
}
