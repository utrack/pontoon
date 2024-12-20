package httpinoapi

import "reflect"

// Option configures handler registration
type Option func(*options)

type options struct {
	tags        []string
	inputType   reflect.Type
	outputType  reflect.Type
	description string
}

// WithTags adds OpenAPI tags to the handler
func WithTags(tags ...string) Option {
	return func(o *options) {
		o.tags = append(o.tags, tags...)
	}
}

// WithInputStruct specifies the input struct for request parameters
func WithInputStruct(inputStruct any) Option {
	return func(o *options) {
		o.inputType = reflect.TypeOf(inputStruct)
	}
}

// WithOutputStruct specifies the output struct for responses
func WithOutputStruct(outputStruct any) Option {
	return func(o *options) {
		o.outputType = reflect.TypeOf(outputStruct)
	}
}

// WithDescription adds a description to the handler
func WithDescription(desc string) Option {
	return func(o *options) {
		o.description = desc
	}
}

func defaultOptions() *options {
	return &options{
		tags: make([]string, 0),
	}
}
