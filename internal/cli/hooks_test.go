package cli

import "testing"

func TestParseHookTemplateRef(t *testing.T) {
	t.Parallel()

	language, projectType, err := parseHookTemplateRef("go/service")
	if err != nil {
		t.Fatalf("parseHookTemplateRef returned error: %v", err)
	}

	if language != "go" || projectType != "service" {
		t.Fatalf("unexpected result: %s/%s", language, projectType)
	}
}

func TestParseHookTemplateRefRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	if _, _, err := parseHookTemplateRef("goservice"); err == nil {
		t.Fatal("expected error for invalid template reference")
	}
}
