package httpinmeditate

import (
	"reflect"
	"strings"
)

// fieldInfo contains parsed httpin field information
type fieldInfo struct {
	Name string
	OriginalName string
	Type reflect.Type
	In   string
	Tags map[string]string
}

// parseField extracts httpin-related information from a struct field
func parseField(field reflect.StructField, isInBody bool) (*fieldInfo, error) {
	info := &fieldInfo{
		Name: field.Name,
		OriginalName: field.Name,
		Type: field.Type,
		Tags: make(map[string]string),
	}

	if isInBody {
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Tag.Get("xml")
		}
		if tag == "" {
			return info, nil
		}
		tvs := strings.Split(tag, ",")
		if len(tvs) == 0 {
			return nil, nil
		}
		if tvs[0] == "-" {
			return nil, nil
		}
		info.Name = tvs[0]
		return info, nil
	}

	tag := field.Tag.Get("httpin")
	if tag == "" {
		return info, nil // not a httpin field
	}

	// Parse httpin tag
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)

		k := strings.TrimSpace(kv[0])
		v := ""

		if len(kv) > 1 {
			v = strings.TrimSpace(kv[1])
			// form/header/json/...
			v = strings.Split(v, ",")[0]
			info.In = k
			if k != "body" {
				info.Name = v
			}
		}
		if isValidationTag(k) {
			continue // skip validation tags
		}
		info.Tags[k] = v
	}

	return info, nil
}

// isValidationTag returns true if the tag is a validation tag
func isValidationTag(tag string) bool {
	switch tag {
	case "required", "default", "nonzero":
		return true
	default:
		return false
	}
}
