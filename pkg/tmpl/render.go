package tmpl

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

// capitalise returns the string with the first rune capitalised
func capitalise(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// ToPascalCase converts a string like "hello world" or "hello-world" to "HelloWorld"
func ToPascalCase(s string) string {
	// Replace non-letter characters with space
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")

	words := strings.Fields(s)
	for i, w := range words {
		words[i] = capitalise(w) // capitalise first letter
	}
	return strings.Join(words, "")
}

// ToKebabCase converts a string like "HelloWorld" or "hello world" to "hello-world"
func ToKebabCase(s string) string {
	// Replace underscores with hyphens and spaces with hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Handle camelCase and PascalCase by inserting hyphens before capital letters
	// and converting to lowercase
	var result bytes.Buffer
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if previous character is already a hyphen
			if s[i-1] != '-' {
				result.WriteRune('-')
			}
			result.WriteRune(r - 'A' + 'a') // Convert to lowercase
		} else if r >= 'A' && r <= 'Z' {
			// First character, just convert to lowercase
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}

	// Remove any double hyphens and trim
	kebab := strings.TrimSpace(result.String())
	for strings.Contains(kebab, "--") {
		kebab = strings.ReplaceAll(kebab, "--", "-")
	}

	return kebab
}

// ToTitleCase converts a string like "hello-world" or "hello_world" to "Hello World"
func ToTitleCase(s string) string {
	// Replace non-letter characters with space
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")

	words := strings.Fields(s)
	for i, w := range words {
		words[i] = capitalise(w) // capitalise first letter
	}
	return strings.Join(words, " ")
}

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

	// Define template functions
	funcMap := template.FuncMap{
		"pascal": ToPascalCase, // Add the PascalCase function as "pascal"
		"kebab":  ToKebabCase,  // Add the KebabCase function as "kebab"
		"title":  ToTitleCase,  // Add the TitleCase function as "title"
	}

	// Parse and execute the modified template with function map
	tmpl, err := template.New("safe").Funcs(funcMap).Parse(result)
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
		if v.Kind() == reflect.Pointer {
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
