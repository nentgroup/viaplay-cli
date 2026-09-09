package template

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

// ValidateManifest checks a manifest for structural problems that would
// otherwise only surface later — mid-scaffold, or not at all if the affected
// option/variable is never actually resolved. An empty/nil result means the
// manifest is valid. A nil manifest (no manifest file present) has no issues.
func ValidateManifest(manifest *Manifest) []string {
	if manifest == nil {
		return nil
	}

	var issues []string

	if !manifest.IsSupported() {
		issues = append(issues, fmt.Sprintf(
			"schema %d is not supported by this version of vip (max supported: 2)", manifest.Schema))
	}

	for i, author := range manifest.Metadata.Authors {
		if author.Email == "" {
			continue
		}
		if _, err := mail.ParseAddress(author.Email); err != nil {
			issues = append(issues, fmt.Sprintf(
				"metadata.authors[%d]: invalid email %q", i, author.Email))
		}
	}

	seenKeys := make(map[string]string) // key -> label of first occurrence, to catch duplicates

	for i, opt := range manifest.Options {
		label := fmt.Sprintf("options[%d]", i)
		if opt.Key == "" {
			issues = append(issues, fmt.Sprintf("%s: key is required", label))
		} else {
			label = fmt.Sprintf("option %q", opt.Key)
			issues = append(issues, checkDuplicateKey(seenKeys, opt.Key, label)...)
		}

		switch strings.ToLower(opt.Type) {
		case "bool", "boolean":
		case "select":
			issues = append(issues, validateSelectChoices(label, opt.Choices)...)
		default:
			issues = append(issues, fmt.Sprintf(
				"%s: unrecognised type %q (expected bool, boolean, or select)", label, opt.Type))
		}
	}

	for i, v := range manifest.Variables {
		label := fmt.Sprintf("variables[%d]", i)
		if v.Key == "" {
			issues = append(issues, fmt.Sprintf("%s: key is required", label))
		} else {
			label = fmt.Sprintf("variable %q", v.Key)
			issues = append(issues, checkDuplicateKey(seenKeys, v.Key, label)...)
		}

		if v.Validate.Pattern != "" {
			if _, err := regexp.Compile(v.Validate.Pattern); err != nil {
				issues = append(issues, fmt.Sprintf("%s: invalid validate.pattern: %v", label, err))
			}
		}
	}

	return issues
}

// checkDuplicateKey records key's first occurrence label and reports an issue
// if it was already seen (declared by both an option and a variable, or twice
// within the same list).
func checkDuplicateKey(seenKeys map[string]string, key, label string) []string {
	if prior, ok := seenKeys[key]; ok {
		return []string{fmt.Sprintf("%s: duplicate key (already used by %s)", label, prior)}
	}
	seenKeys[key] = label
	return nil
}

// validateSelectChoices checks that a "select" option declares at least one
// choice, each with a non-empty value.
func validateSelectChoices(label string, choices []ManifestChoice) []string {
	var issues []string
	if len(choices) == 0 {
		issues = append(issues, fmt.Sprintf("%s: type select requires at least one choice", label))
	}
	for i, choice := range choices {
		if choice.Value == "" {
			issues = append(issues, fmt.Sprintf("%s: choices[%d] has an empty value", label, i))
		}
	}
	return issues
}
