package template

import (
	"strings"
	"testing"
)

const (
	kebabCasePattern    = "^[a-z-]+$"
	shortNameKey        = "shortName"
	invalidShortName    = "MyService"
	validShortName      = "my-service"
	kebabCaseTypeString = "string"
	kebabCaseMessage    = "must be lowercase kebab-case"
	regionKey           = "region"
	boolTypeString      = "bool"
	sqsKey              = "sqs"
	selectTypeString    = "select"
	authorName          = "Jane Doe"
)

func TestValidateManifestVariable(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		message string
		value   any
		wantErr bool
		errMsg  string
	}{
		{name: "no pattern always passes", pattern: "", value: "anything", wantErr: false},
		{name: "matching pattern passes", pattern: kebabCasePattern, value: validShortName, wantErr: false},
		{
			name: "non-matching pattern fails with custom message", pattern: kebabCasePattern,
			message: kebabCaseMessage, value: invalidShortName, wantErr: true,
			errMsg: kebabCaseMessage,
		},
		{
			name: "non-matching pattern fails with generic message", pattern: kebabCasePattern,
			value: invalidShortName, wantErr: true,
			errMsg: `value "MyService" does not match pattern "^[a-z-]+$"`,
		},
		{name: "non-string value is stringified before matching", pattern: "^[0-9]+$", value: 42, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variable := ManifestVariable{
				Key:      shortNameKey,
				Validate: ManifestValidate{Pattern: tt.pattern, Message: tt.message},
			}
			err := validateManifestVariable(variable, tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Fatalf("expected error %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestResolveManifestSelections_SetOverrideValidation(t *testing.T) {
	manifest := &Manifest{
		Variables: []ManifestVariable{
			{
				Key:      shortNameKey,
				Type:     kebabCaseTypeString,
				Required: true,
				Validate: ManifestValidate{Pattern: kebabCasePattern, Message: kebabCaseMessage},
			},
		},
	}

	t.Run("valid override is accepted", func(t *testing.T) {
		vars := &Variables{}
		err := ResolveManifestSelections(manifest, vars, []string{shortNameKey + "=" + validShortName}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vars.Features[shortNameKey] != validShortName {
			t.Fatalf("expected shortName to be set, got %v", vars.Features[shortNameKey])
		}
	})

	t.Run("invalid override is rejected", func(t *testing.T) {
		vars := &Variables{}
		err := ResolveManifestSelections(manifest, vars, []string{shortNameKey + "=" + invalidShortName}, true)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})
}

func TestResolveManifestSelections_NoInputDefaultValidation(t *testing.T) {
	t.Run("invalid default is rejected in no-input mode", func(t *testing.T) {
		manifest := &Manifest{
			Variables: []ManifestVariable{
				{
					Key:      shortNameKey,
					Type:     kebabCaseTypeString,
					Default:  invalidShortName,
					Validate: ManifestValidate{Pattern: kebabCasePattern},
				},
			},
		}
		vars := &Variables{}
		err := ResolveManifestSelections(manifest, vars, nil, true)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})

	t.Run("valid default is accepted in no-input mode", func(t *testing.T) {
		manifest := &Manifest{
			Variables: []ManifestVariable{
				{
					Key:      shortNameKey,
					Type:     kebabCaseTypeString,
					Default:  validShortName,
					Validate: ManifestValidate{Pattern: kebabCasePattern},
				},
			},
		}
		vars := &Variables{}
		err := ResolveManifestSelections(manifest, vars, nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vars.Features[shortNameKey] != validShortName {
			t.Fatalf("expected shortName to be set, got %v", vars.Features[shortNameKey])
		}
	})
}

func TestResolveManifestSelections_NoInputCollectsAllMissingRequired(t *testing.T) {
	manifest := &Manifest{
		Options: []ManifestOption{
			{Key: sqsKey, Type: boolTypeString, Required: true},
		},
		Variables: []ManifestVariable{
			{Key: shortNameKey, Required: true},
			{Key: regionKey, Required: true},
		},
	}
	vars := &Variables{}
	err := ResolveManifestSelections(manifest, vars, nil, true)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	for _, want := range []string{sqsKey, shortNameKey, regionKey} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %q, got: %v", want, err)
		}
	}
}

func TestResolveManifestSelections_NoInputPartialSetStillReportsRemaining(t *testing.T) {
	manifest := &Manifest{
		Variables: []ManifestVariable{
			{Key: shortNameKey, Type: kebabCaseTypeString, Required: true},
			{Key: regionKey, Required: true},
		},
	}
	vars := &Variables{}
	err := ResolveManifestSelections(manifest, vars, []string{shortNameKey + "=" + validShortName}, true)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if strings.Contains(err.Error(), shortNameKey) {
		t.Fatalf("did not expect error to mention already-provided %q, got: %v", shortNameKey, err)
	}
	if !strings.Contains(err.Error(), regionKey) {
		t.Fatalf("expected error to mention %q, got: %v", regionKey, err)
	}
}

func TestResolveManifestSelections_RejectsInvalidManifestBeforeResolving(t *testing.T) {
	manifest := &Manifest{
		Options: []ManifestOption{
			{Key: sqsKey, Type: "not-a-real-type"},
		},
	}
	vars := &Variables{}
	err := ResolveManifestSelections(manifest, vars, nil, true)
	if err == nil {
		t.Fatalf("expected an error for an invalid manifest, got nil")
	}
	if !strings.Contains(err.Error(), "unrecognised type") {
		t.Fatalf("expected error to mention the manifest validation issue, got: %v", err)
	}
}
