package template

import (
	"strings"
	"testing"
)

func TestValidateManifest_NilManifestHasNoIssues(t *testing.T) {
	t.Parallel()

	if issues := ValidateManifest(nil); issues != nil {
		t.Fatalf("expected no issues for a nil manifest, got: %v", issues)
	}
}

func TestValidateManifest_ValidManifestHasNoIssues(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Schema: 2,
		Options: []ManifestOption{
			{Key: sqsKey, Type: boolTypeString},
			{Key: regionKey, Type: selectTypeString, Choices: []ManifestChoice{{Value: "eu"}, {Value: "us"}}},
		},
		Variables: []ManifestVariable{
			{Key: shortNameKey, Type: kebabCaseTypeString, Validate: ManifestValidate{Pattern: kebabCasePattern}},
		},
	}

	if issues := ValidateManifest(manifest); len(issues) != 0 {
		t.Fatalf("expected no issues, got: %v", issues)
	}
}

func TestValidateManifest_UnsupportedSchema(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{Schema: 3}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "schema 3 is not supported") {
		t.Fatalf("expected an unsupported schema issue, got: %v", issues)
	}
}

func TestValidateManifest_MissingOptionKey(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{Options: []ManifestOption{{Type: boolTypeString}}}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "key is required") {
		t.Fatalf("expected a missing key issue, got: %v", issues)
	}
}

func TestValidateManifest_UnrecognisedOptionType(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{Options: []ManifestOption{{Key: sqsKey, Type: "yesno"}}}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, `unrecognised type "yesno"`) {
		t.Fatalf("expected an unrecognised type issue, got: %v", issues)
	}
}

func TestValidateManifest_SelectWithNoChoices(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{Options: []ManifestOption{{Key: "region", Type: selectTypeString}}}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "requires at least one choice") {
		t.Fatalf("expected a missing choices issue, got: %v", issues)
	}
}

func TestValidateManifest_SelectWithEmptyChoiceValue(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Options: []ManifestOption{
			{Key: "region", Type: selectTypeString, Choices: []ManifestChoice{{Value: "eu"}, {Label: "US"}}},
		},
	}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "choices[1] has an empty value") {
		t.Fatalf("expected an empty choice value issue, got: %v", issues)
	}
}

func TestValidateManifest_DuplicateKeyAcrossOptionsAndVariables(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Options:   []ManifestOption{{Key: shortNameKey, Type: boolTypeString}},
		Variables: []ManifestVariable{{Key: shortNameKey}},
	}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "duplicate key") {
		t.Fatalf("expected a duplicate key issue, got: %v", issues)
	}
}

func TestValidateManifest_InvalidValidatePattern(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Variables: []ManifestVariable{
			{Key: shortNameKey, Validate: ManifestValidate{Pattern: "["}},
		},
	}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "invalid validate.pattern") {
		t.Fatalf("expected an invalid pattern issue, got: %v", issues)
	}
}

func TestValidateManifest_InvalidAuthorEmail(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Metadata: ManifestMetadata{
			Authors: []ManifestAuthor{{Name: authorName, Email: "not-an-email"}},
		},
	}
	issues := ValidateManifest(manifest)
	if !containsIssue(issues, "invalid email") {
		t.Fatalf("expected an invalid email issue, got: %v", issues)
	}
}

func TestValidateManifest_ValidAuthorEmailHasNoIssues(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Metadata: ManifestMetadata{
			Authors: []ManifestAuthor{{Name: authorName, Email: "jane@example.com"}},
		},
	}
	issues := ValidateManifest(manifest)
	if containsIssue(issues, "invalid email") {
		t.Fatalf("expected no invalid email issue, got: %v", issues)
	}
}

func TestValidateManifest_AuthorWithoutEmailHasNoIssues(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Metadata: ManifestMetadata{
			Authors: []ManifestAuthor{{Name: authorName}},
		},
	}
	issues := ValidateManifest(manifest)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got: %v", issues)
	}
}

func containsIssue(issues []string, substr string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, substr) {
			return true
		}
	}
	return false
}
