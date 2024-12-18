package httpinmeditate

import (
	"encoding"
	"fmt"
	"reflect"
	"time"

	base "github.com/pb33f/libopenapi/datamodel/high/base"
)

// primitiveTypeToSchema converts a primitive Go type to an OpenAPI schema
func primitiveTypeToSchema(t reflect.Type) (*base.Schema, error) {
	debugLog("primitiveTypeToSchema: processing type %v (kind: %v, name: %q)", t, t.Kind(), t.Name())

	// Handle special types first
	switch t {
	case reflect.TypeOf(time.Time{}):
		debugLog("-> matched time.Time")
		return &base.Schema{
			Type:   []string{"string"},
			Format: "date-time",
		}, nil
	case reflect.TypeOf([]byte(nil)):
		debugLog("-> matched []byte")
		return &base.Schema{
			Type:   []string{"string"},
			Format: "binary",
		}, nil
	}

	// Handle time.Time-like types and text marshalers for named types
	if t.Name() != "" {
		debugLog("-> processing named type %q", t.Name())
		// Handle time.Time-like types
		if t.ConvertibleTo(reflect.TypeOf(time.Time{})) {
			debugLog("  -> convertible to time.Time")
			return &base.Schema{
				Type:   []string{"string"},
				Format: "date-time",
			}, nil
		}

		// Handle custom types that implement encoding.TextMarshaler
		if t.Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()) {
			debugLog("  -> implements TextMarshaler")
			return &base.Schema{Type: []string{"string"}}, nil
		}
	}

	// Handle types based on their kind
	switch t.Kind() {
	case reflect.Bool:
		debugLog("-> matched bool")
		return &base.Schema{Type: []string{"boolean"}}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		debugLog("-> matched integer")
		return &base.Schema{Type: []string{"integer"}}, nil
	case reflect.Float32, reflect.Float64:
		debugLog("-> matched float")
		return &base.Schema{Type: []string{"number"}}, nil
	case reflect.String:
		debugLog("-> matched string")
		return &base.Schema{Type: []string{"string"}}, nil
	case reflect.Array, reflect.Slice:
		debugLog(" -> non-primitive slice, ret nil")
		return nil, nil
	case reflect.Map:
		debugLog(" -> non-primitive map, ret nil")
		return nil, nil
	case reflect.Ptr:
		debugLog(" -> non-primitive ptr, ret nil")
		return nil, nil
	case reflect.Struct:
		debugLog(" -> non-primitive struct, ret nil")
		return nil, nil
	}

	debugLog("-> type %v not supported", t)
	return nil, fmt.Errorf("type %v not supported", t)
}
