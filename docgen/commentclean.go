package docgen

import (
	"strings"
	"unicode"
)

// docFromComment returns a cleaned-up comment for a field or method.
func docFromComment(goName string, jsonTag string, comment string) string {

	// foo.Bar -> Bar
	if idx := strings.LastIndex(goName, "."); idx > -1 {
		goName = goName[idx+1:]
	}

	// remove heading FieldName needed by Go specs
	//
	// FooField is a foo field -> is a foo field
	// fooFielD is a foo field -> is a foo field
	if strings.HasPrefix(strings.ToLower(comment), strings.ToLower(goName)) {
		comment = comment[len(goName):]
	}

	comment = strings.Trim(comment, "\n\r \t")

	comment = strings.TrimPrefix(comment, "is ")

	// replace any other goName occurences with jsonTag if it's there
	if jsonTag != "" {
		comment = strings.ReplaceAll(comment, goName, jsonTag)
	}

	// capitalize first letter
	if len(comment) > 0 {
		r := []rune(comment)
		r[0] = unicode.ToUpper(r[0])
		comment = string(r)
	}
	return comment
}
