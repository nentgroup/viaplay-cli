package tmpl

import (
	"testing"
)

const (
	testProjectName = "TestProject"
	testKey1        = "key1"
)

func TestRenderWithLiteralUnknowns(t *testing.T) {
	// Test data structure with various field types
	type TestData struct {
		Name       string
		Count      int
		IsActive   bool
		Tags       []string
		Properties map[string]string
		Nested     struct {
			Value string
		}
	}

	// Create test data with realistic values
	data := TestData{
		Name:     testProjectName,
		Count:    42,
		IsActive: true,
		Tags:     []string{"go", "template", "test"},
		Properties: map[string]string{
			testKey1: "value1",
			"key2":   "value2",
		},
		Nested: struct {
			Value string
		}{
			Value: "NestedValue",
		},
	}

	// Test cases
	tests := []struct {
		name           string
		templateString string
		data           interface{}
		expected       string
		expectError    bool
	}{
		// Standard Go template features (these should work normally)
		{
			name:           "Simple variable",
			templateString: "Project: {{ .Name }}",
			data:           data,
			expected:       "Project: TestProject",
			expectError:    false,
		},
		{
			name:           "Integer variable",
			templateString: "Count: {{ .Count }}",
			data:           data,
			expected:       "Count: 42",
			expectError:    false,
		},
		{
			name:           "Boolean variable",
			templateString: "Active: {{ .IsActive }}",
			data:           data,
			expected:       "Active: true",
			expectError:    false,
		},
		{
			name:           "Nested field",
			templateString: "Nested: {{ .Nested.Value }}",
			data:           data,
			expected:       "Nested: NestedValue",
			expectError:    false,
		},
		{
			name:           "Array access using index function",
			templateString: "First tag: {{ index .Tags 0 }}",
			data:           data,
			expected:       "First tag: go",
			expectError:    false,
		},
		{
			name:           "Map access",
			templateString: "Property: {{ .Properties.key1 }}",
			data:           data,
			expected:       "Property: value1",
			expectError:    false,
		},
		{
			name:           "Conditional",
			templateString: "{{ if .IsActive }}Active{{ else }}Inactive{{ end }}",
			data:           data,
			expected:       "Active",
			expectError:    false,
		},
		{
			name:           "Range over array",
			templateString: "Tags: {{ range .Tags }}{{ . }} {{ end }}",
			data:           data,
			expected:       "Tags: go template test ",
			expectError:    false,
		},
		{
			name:           "Multiple variables",
			templateString: "{{ .Name }} has {{ .Count }} items",
			data:           data,
			expected:       "TestProject has 42 items",
			expectError:    false,
		},
		{
			name:           "Whitespace control",
			templateString: "{{- .Name -}}",
			data:           data,
			expected:       testProjectName,
			expectError:    false,
		},

		// Missing variables (our special case - should preserve the template expressions)
		{
			name:           "Missing variable should be preserved",
			templateString: "Missing: {{ .MissingField }}",
			data:           data,
			expected:       "Missing: {{ .MissingField }}",
			expectError:    false,
		},
		{
			name:           "Missing nested field should be preserved",
			templateString: "Missing nested: {{ .Nested.MissingField }}",
			data:           data,
			expected:       "Missing nested: {{ .Nested.MissingField }}",
			expectError:    false,
		},
		{
			name:           "Out of bounds array index should be preserved",
			templateString: "Out of bounds: {{ index .Tags 10 }}",
			data:           data,
			expected:       "Out of bounds: {{ index .Tags 10 }}",
			expectError:    true,
		},
		{
			name:           "Non-existent map key should be preserved",
			templateString: "Missing key: {{ .Properties.nonexistent }}",
			data:           data,
			expected:       "Missing key: {{ .Properties.nonexistent }}",
			expectError:    false,
		},
		{
			name:           "Mixed existing and missing fields",
			templateString: "{{ .Name }} has {{ .MissingField }} and {{ .Count }}",
			data:           data,
			expected:       "TestProject has {{ .MissingField }} and 42",
			expectError:    false,
		},

		// Edge cases
		{
			name:           "Empty template",
			templateString: "",
			data:           data,
			expected:       "",
			expectError:    false,
		},
		{
			name:           "Template with only text",
			templateString: "Just text, no variables",
			data:           data,
			expected:       "Just text, no variables",
			expectError:    false,
		},
		{
			name:           "Invalid template syntax",
			templateString: "Invalid {{ .Name syntax",
			data:           data,
			expected:       "",
			expectError:    true,
		},
		{
			name:           "Nil data",
			templateString: "Nil: {{ .Field }}",
			data:           nil,
			expected:       "Nil: {{ .Field }}",
			expectError:    false,
		},
		{
			name:           "Empty struct data",
			templateString: "Empty: {{ .Field }}",
			data:           struct{}{},
			expected:       "Empty: {{ .Field }}",
			expectError:    false,
		},
	}

	// Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderWithLiteralUnknowns(tt.templateString, tt.data)

			// Check if we got an error when we expected one
			if (err != nil) != tt.expectError {
				t.Errorf("Error expectation mismatch - got error: %v, expectError: %v", err, tt.expectError)
				return
			}

			// Skip result check if we expected an error
			if err != nil {
				return
			}

			// Compare the result with the expected output
			if result != tt.expected {
				t.Errorf("Result mismatch\nexpected: %q\ngot:      %q", tt.expected, result)
			}
		})
	}
}

// Test the fieldExists helper function
func TestFieldExists(t *testing.T) {
	// Test data structure
	type TestData struct {
		Name       string
		Count      int
		Tags       []string
		Properties map[string]string
		Nested     struct {
			Value string
		}
	}

	// Create test data
	data := TestData{
		Name:  testProjectName,
		Count: 42,
		Tags:  []string{"go", "template", "test"},
		Properties: map[string]string{
			testKey1: "value1",
			"key2":   "value2",
		},
		Nested: struct {
			Value string
		}{
			Value: "NestedValue",
		},
	}

	// Test cases for fieldExists function
	tests := []struct {
		name     string
		data     interface{}
		path     []string
		expected bool
	}{
		{
			name:     "Simple field exists",
			data:     data,
			path:     []string{"Name"},
			expected: true,
		},
		{
			name:     "Simple field does not exist",
			data:     data,
			path:     []string{"MissingField"},
			expected: false,
		},
		{
			name:     "Nested field exists",
			data:     data,
			path:     []string{"Nested", "Value"},
			expected: true,
		},
		{
			name:     "Nested field does not exist",
			data:     data,
			path:     []string{"Nested", "MissingField"},
			expected: false,
		},
		{
			name:     "Array element out of bounds",
			data:     data,
			path:     []string{"Tags", "10"},
			expected: false,
		},
		{
			name:     "Map key exists",
			data:     data,
			path:     []string{"Properties", testKey1},
			expected: true,
		},
		{
			name:     "Map key does not exist",
			data:     data,
			path:     []string{"Properties", "nonexistent"},
			expected: false,
		},
		{
			name:     "Nil data",
			data:     nil,
			path:     []string{"Field"},
			expected: false,
		},
		{
			name:     "Empty path",
			data:     data,
			path:     []string{},
			expected: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldExists(tt.data, tt.path)
			if result != tt.expected {
				t.Errorf("fieldExists(%v, %v) = %v; expected %v",
					tt.data, tt.path, result, tt.expected)
			}
		})
	}
}

func TestTemplateCaseHelpers(t *testing.T) {
	funcs := FuncMap()

	lower, ok := funcs["lower"].(func(string) string)
	if !ok {
		t.Fatal("lower helper is not registered")
	}
	if got := lower("Hello World"); got != "hello world" {
		t.Fatalf("lower helper = %q, want %q", got, "hello world")
	}

	upper, ok := funcs["upper"].(func(string) string)
	if !ok {
		t.Fatal("upper helper is not registered")
	}
	if got := upper("Hello World"); got != "HELLO WORLD" {
		t.Fatalf("upper helper = %q, want %q", got, "HELLO WORLD")
	}

	camel, ok := funcs["camel"].(func(string) string)
	if !ok {
		t.Fatal("camel helper is not registered")
	}
	if got := camel("my-service_name"); got != "myServiceName" {
		t.Fatalf("camel helper = %q, want %q", got, "myServiceName")
	}

	snake, ok := funcs["snake"].(func(string) string)
	if !ok {
		t.Fatal("snake helper is not registered")
	}
	if got := snake("MyService"); got != "my_service" {
		t.Fatalf("snake helper = %q, want %q", got, "my_service")
	}
}

func TestTemplateSlugAndReplaceHelpers(t *testing.T) {
	funcs := FuncMap()
	const expectedSlug = "my-service"
	const expectedResult = "my-service"

	slug, ok := funcs["slug"].(func(string) string)
	if !ok {
		t.Fatal("slug helper is not registered")
	}
	if got := slug("My Service!!!"); got != expectedSlug {
		t.Fatalf("slug helper = %q, want %q", got, expectedSlug)
	}

	replace, ok := funcs["replace"].(func(string, string, ...string) string)
	if !ok {
		t.Fatal("replace helper is not registered")
	}
	if got := replace(".", "-", "my.service"); got != expectedResult {
		t.Fatalf("replace helper = %q, want %q", got, expectedResult)
	}
}

func TestTemplateFallbackAndCollectionHelpers(t *testing.T) {
	funcs := FuncMap()

	defaultValue, ok := funcs["default"].(func(string, ...string) string)
	if !ok {
		t.Fatal("default helper is not registered")
	}
	if got := defaultValue("fallback", ""); got != "fallback" {
		t.Fatalf("default helper = %q, want %q", got, "fallback")
	}

	coalesce, ok := funcs["coalesce"].(func(...string) string)
	if !ok {
		t.Fatal("coalesce helper is not registered")
	}
	if got := coalesce("", "value"); got != "value" {
		t.Fatalf("coalesce helper = %q, want %q", got, "value")
	}

	join, ok := funcs["join"].(func(string, interface{}) string)
	if !ok {
		t.Fatal("join helper is not registered")
	}
	if got := join(",", []string{"a", "b"}); got != "a,b" {
		t.Fatalf("join helper = %q, want %q", got, "a,b")
	}

	split, ok := funcs["split"].(func(string, string) []string)
	if !ok {
		t.Fatal("split helper is not registered")
	}
	if got := split(",", "a,b,c"); len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("split helper = %#v, want [a b c]", got)
	}
}

func TestTemplateTrimHelpers(t *testing.T) {
	funcs := FuncMap()
	const expectedResult = "my-service"

	trim, ok := funcs["trim"].(func(string) string)
	if !ok {
		t.Fatal("trim helper is not registered")
	}
	if got := trim("  hi  "); got != "hi" {
		t.Fatalf("trim helper = %q, want %q", got, "hi")
	}

	trimSuffix, ok := funcs["trimSuffix"].(func(string, string) string)
	if !ok {
		t.Fatal("trimSuffix helper is not registered")
	}
	if got := trimSuffix("-", "my-service-"); got != expectedResult {
		t.Fatalf("trimSuffix helper = %q, want %q", got, expectedResult)
	}

	trimPrefix, ok := funcs["trimPrefix"].(func(string, string) string)
	if !ok {
		t.Fatal("trimPrefix helper is not registered")
	}
	if got := trimPrefix("pre-", "pre-my-service"); got != expectedResult {
		t.Fatalf("trimPrefix helper = %q, want %q", got, expectedResult)
	}
}

func TestTemplateIDHelpers(t *testing.T) {
	funcs := FuncMap()

	uuid, ok := funcs["uuid"].(func() string)
	if !ok {
		t.Fatal("uuid helper is not registered")
	}
	if got := uuid(); len(got) != 36 {
		t.Fatalf("uuid helper len = %d, want 36", len(got))
	}

	ulid, ok := funcs["ulid"].(func() string)
	if !ok {
		t.Fatal("ulid helper is not registered")
	}
	if got := ulid(); len(got) != 26 {
		t.Fatalf("ulid helper len = %d, want 26", len(got))
	}
}
