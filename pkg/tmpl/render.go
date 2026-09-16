package tmpl

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/oklog/ulid/v2"
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

// ToCamelCase converts a string like "my-service" or "my_service" to "myService".
func ToCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(parts) == 0 {
		return ""
	}

	var b strings.Builder
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(trimmed))
			continue
		}
		b.WriteString(capitalise(strings.ToLower(trimmed)))
	}
	if b.Len() == 0 {
		return ""
	}
	return b.String()
}

// ToSnakeCase converts a string like "MyService" or "my-service" to "my_service".
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	for i, r := range s {
		switch {
		case unicode.IsUpper(r):
			if i > 0 {
				prev := rune(s[i-1])
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteRune('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case b.Len() > 0 && b.String()[b.Len()-1] != '_':
			b.WriteRune('_')
		}
	}

	out := strings.Trim(b.String(), "_")
	out = strings.ReplaceAll(out, "__", "_")
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	return out
}

// ToSlug converts a string into a lowercase, URL-safe slug.
func ToSlug(s string) string {
	if s == "" {
		return ""
	}

	lower := strings.ToLower(s)
	lower = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '-'
	}, lower)
	lower = strings.Trim(lower, "-")
	for strings.Contains(lower, "--") {
		lower = strings.ReplaceAll(lower, "--", "-")
	}
	return lower
}

// DefaultString returns the first non-empty string, falling back to the default value.
func DefaultString(defaultValue string, values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return defaultValue
}

// CoalesceString returns the first non-empty string from the provided values.
func CoalesceString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// JoinStrings joins a slice of strings using the provided separator.
func JoinStrings(sep string, values interface{}) string {
	switch v := values.(type) {
	case []string:
		return strings.Join(v, sep)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, sep)
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

// SplitString splits a string on the provided separator.
func SplitString(sep, value string) []string {
	if value == "" {
		return []string{""}
	}
	return strings.Split(value, sep)
}

// ReplaceString replaces every occurrence of old in value with new.
func ReplaceString(old, new string, values ...string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.ReplaceAll(values[0], old, new)
}

// TrimPrefixString removes the prefix from a value.
func TrimPrefixString(prefix, value string) string {
	return strings.TrimPrefix(value, prefix)
}

// TrimSuffixString removes the suffix from a value.
func TrimSuffixString(suffix, value string) string {
	return strings.TrimSuffix(value, suffix)
}

// GenerateUUID returns a RFC 4122 version 4 UUID.
func GenerateUUID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return ""
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], raw[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], raw[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], raw[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], raw[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], raw[10:16])
	return string(buf)
}

// FuncMap returns the shared helper functions available to templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"pascal":     ToPascalCase,
		"kebab":      ToKebabCase,
		"title":      ToTitleCase,
		"camel":      ToCamelCase,
		"snake":      ToSnakeCase,
		"slug":       ToSlug,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"replace":    ReplaceString,
		"default":    DefaultString,
		"coalesce":   CoalesceString,
		"join":       JoinStrings,
		"split":      SplitString,
		"trim":       strings.TrimSpace,
		"trimSuffix": TrimSuffixString,
		"trimPrefix": TrimPrefixString,
		"uuid":       func() string { return GenerateUUID() },
		"uuidv4":     func() string { return GenerateUUID() },
		"ulid":       func() string { return ulid.Make().String() },
	}
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

	// Reuse the shared helper set so file content rendering supports the same
	// functions as file/directory name rendering (pascal/kebab/title, plus
	// case helpers and ID generators like uuid/uuidv4/ulid).
	funcMap := FuncMap()

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
