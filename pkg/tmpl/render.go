package tmpl

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// RenderWithLiteralUnknowns processes a Go template string so that
// missing variables remain literally in the output, instead of causing errors.
func RenderWithLiteralUnknowns(templateString string, data interface{}) (string, error) {
	// Regex to find all {{ .Field }} or nested like {{ .User.Name }}
	fieldPattern := regexp.MustCompile(`{{\s*\.([A-Za-z0-9_\.]+)\s*}}`)

	// Keep track of which expressions to preserve literally
	literalExpressions := make(map[string]bool)

	// Find all field references and check if they exist
	fieldPattern.ReplaceAllStringFunc(templateString, func(match string) string {
		if literalExpressions[match] {
			return match // already marked
		}

		parts := fieldPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		pathParts := strings.Split(parts[1], ".")

		// Skip if any path part is numeric (we do not support .Tags.0 style here)
		for _, part := range pathParts {
			if _, err := strconv.Atoi(part); err == nil {
				return match
			}
		}

		// Check if the field exists in the data struct/map
		if !fieldExists(data, pathParts) {
			literalExpressions[match] = true
		}

		return match
	})

	// Replace missing fields with backtick-escaped versions so they render literally
	result := templateString
	for expr := range literalExpressions {
		escaped := "{{`" + expr + "`}}"
		result = strings.ReplaceAll(result, expr, escaped)
	}

	// Parse and execute the modified template
	tmpl, err := template.New("safe").Parse(result)
	if err != nil {
		return "", fmt.Errorf("failed to parse template string: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template string: %w", err)
	}

	return buf.String(), nil
}

// fieldExists checks if a nested field/index exists in struct/map/slice/array
func fieldExists(data interface{}, path []string) bool {
	v := reflect.ValueOf(data)

	for _, p := range path {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return false
			}
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			f := v.FieldByName(p)
			if !f.IsValid() {
				return false
			}
			v = f

		case reflect.Map:
			val := v.MapIndex(reflect.ValueOf(p))
			if !val.IsValid() {
				return false
			}
			v = val

		case reflect.Slice, reflect.Array:
			// Since we don't support numeric indices in path, this should never happen here
			return false

		default:
			return false
		}
	}

	return true
}
