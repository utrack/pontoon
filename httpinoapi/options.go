package httpinoapi

import "reflect"

// Option configures handler registration
type Option func(*options)

type options struct {
	tags       []string
	inputType  reflect.Type
	outputType reflect.Type
}

// WithTags adds OpenAPI tags to the handler
func WithTags(tags ...string) Option {
	return func(o *options) {
		o.tags = append(o.tags, tags...)
	}
}

// WithInputType specifies the input type for request parameters
func WithInputType(inputType reflect.Type) Option {
	return func(o *options) {
		o.inputType = inputType
	}
}

// WithInputStruct specifies the input struct for request parameters
func WithInputStruct(inputStruct any) Option {
	return func(o *options) {
		o.inputType = reflect.TypeOf(inputStruct)
	}
}

// WithOutputType specifies the output type for responses
func WithOutputType(outputType reflect.Type) Option {
	return func(o *options) {
		o.outputType = outputType
	}
}

// WithOutputStruct specifies the output struct for responses
func WithOutputStruct(outputStruct any) Option {
	return func(o *options) {
		o.outputType = reflect.TypeOf(outputStruct)
	}
}

func defaultOptions() *options {
	return &options{
		tags: make([]string, 0),
	}
}
