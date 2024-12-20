package httpinoapi

import (
	"fmt"
	"net/http"
	"reflect"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Generator generates OpenAPI documentation from HTTP handlers.
type Generator struct {
	handlers []*handlerInfo

	components *v3.Components
}

type handlerInfo struct {
	verb    string
	path    string
	handler interface{}
	options *options
}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{
		handlers: make([]*handlerInfo, 0),
	}
}

// Operation registers an HTTP handler in the group.
func (g *Generator) Operation(verb string, path string, handler interface{}, opts ...Option) error {
	if err := validateVerb(verb); err != nil {
		return fmt.Errorf("invalid verb: %w", err)
	}

	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if err := validateHandler(handler); err != nil {
		return fmt.Errorf("invalid handler: %w", err)
	}

	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := g.validateDuplicates(verb, path); err != nil {
		return fmt.Errorf("duplicate operation: %w", err)
	}

	g.handlers = append(g.handlers, &handlerInfo{
		verb:    verb,
		path:    path,
		handler: handler,
		options: options,
	})

	return nil
}

func validateVerb(verb string) error {
	switch verb {
	case http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodPatch, http.MethodHead,
		http.MethodOptions, http.MethodTrace:
		return nil
	default:
		return fmt.Errorf("unsupported HTTP verb: %s", verb)
	}
}

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}
	return nil
}

func validateHandler(handler interface{}) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	t := reflect.TypeOf(handler)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}
	return nil
}

func (g *Generator) validateDuplicates(verb, path string) error {
	for _, h := range g.handlers {
		if h.verb == verb && h.path == path {
			return fmt.Errorf("handler already registered for %s %s", verb, path)
		}
	}
	return nil
}
