package template

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

// ResolveManifestSelections resolves manifest prompts and command-line overrides.
func ResolveManifestSelections(manifest *Manifest, vars *Variables, overrides []string, noInput bool) error {
	if manifest == nil {
		return nil
	}
	if vars.Features == nil {
		vars.Features = FeatureSet{}
	}
	for _, raw := range overrides {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set value %q, expected key=value", raw)
		}
		key := strings.TrimSpace(parts[0])
		value := parseSelectionValue(strings.TrimSpace(parts[1]))
		if variable, ok := findManifestVariable(manifest.Variables, key); ok {
			if err := validateManifestVariable(variable, value); err != nil {
				return fmt.Errorf("invalid --set value for %q: %w", key, err)
			}
		}
		vars.SetFeature(key, value)
	}
	for _, opt := range manifest.Options {
		if _, ok := vars.Features[opt.Key]; ok {
			continue
		}
		if noInput {
			if opt.Required {
				return fmt.Errorf("missing required template option %q", opt.Key)
			}
			vars.SetFeature(opt.Key, opt.Default)
			continue
		}
		value, err := promptOption(opt)
		if err != nil {
			return err
		}
		vars.SetFeature(opt.Key, value)
	}
	for _, variable := range manifest.Variables {
		if _, ok := vars.Features[variable.Key]; ok {
			continue
		}
		if noInput {
			if variable.Required {
				return fmt.Errorf("missing required template variable %q", variable.Key)
			}
			if err := validateManifestVariable(variable, variable.Default); err != nil {
				return fmt.Errorf("invalid default value for template variable %q: %w", variable.Key, err)
			}
			vars.SetFeature(variable.Key, variable.Default)
			continue
		}
		value, err := promptVariableWithValidation(variable)
		if err != nil {
			return err
		}
		vars.SetFeature(variable.Key, value)
	}
	return nil
}

func promptOption(opt ManifestOption) (any, error) {
	label := opt.Prompt
	if label == "" {
		label = opt.Key
	}
	switch strings.ToLower(opt.Type) {
	case "bool", "boolean":
		prompt := promptui.Select{
			Label: label,
			Items: []string{"no", "yes"},
			Templates: &promptui.SelectTemplates{
				Selected: fmt.Sprintf(`{{ "✔" | green }} %s: {{ . | faint }}`, label),
			},
		}
		_, result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		return result == "yes", nil
	case "select":
		items := make([]string, 0, len(opt.Choices))
		for _, choice := range opt.Choices {
			if choice.Label != "" {
				items = append(items, choice.Label)
			} else {
				items = append(items, choice.Value)
			}
		}
		if len(items) == 0 {
			items = []string{fmt.Sprint(opt.Default)}
		}
		prompt := promptui.Select{
			Label: label,
			Items: items,
			Templates: &promptui.SelectTemplates{
				Selected: fmt.Sprintf(`{{ "✔" | green }} %s: {{ . | faint }}`, label),
			},
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		if idx >= 0 && idx < len(opt.Choices) {
			return opt.Choices[idx].Value, nil
		}
		return nil, nil
	default:
		prompt := promptui.Prompt{
			Label:   label,
			Default: defaultString(opt.Default),
			Templates: &promptui.PromptTemplates{
				Success: fmt.Sprintf(`{{ "✔" | green }} %s: `, label),
			},
		}
		text, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return opt.Default, nil
		}
		return text, nil
	}
}

func promptVariable(variable ManifestVariable) (any, error) {
	label := variable.Prompt
	if label == "" {
		label = variable.Key
	}
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultString(variable.Default),
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf(`{{ "✔" | green }} %s: `, label),
		},
		Validate: func(input string) error {
			text := strings.TrimSpace(input)
			if text == "" {
				if variable.Required {
					return fmt.Errorf("%s is required", variable.Key)
				}
				return nil
			}
			return validateManifestVariable(variable, text)
		},
	}
	text, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return variable.Default, nil
	}
	return text, nil
}

// promptVariableWithValidation prompts for a manifest variable's value.
// Validation (variable.Validate.Pattern/Message) runs inline as the user
// types, via promptui's Validate hook, so invalid input is rejected before
// it can be submitted.
func promptVariableWithValidation(variable ManifestVariable) (any, error) {
	return promptVariable(variable)
}

// findManifestVariable looks up a declared variable by key.
func findManifestVariable(variables []ManifestVariable, key string) (ManifestVariable, bool) {
	for _, v := range variables {
		if v.Key == key {
			return v, true
		}
	}
	return ManifestVariable{}, false
}

// validateManifestVariable checks value against variable.Validate.Pattern, if
// one is declared. Values are compared as their string representation
// (fmt.Sprint), so this works uniformly for string, numeric, etc. variables.
func validateManifestVariable(variable ManifestVariable, value any) error {
	pattern := variable.Validate.Pattern
	if pattern == "" {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid validate.pattern for template variable %q: %w", variable.Key, err)
	}
	str := fmt.Sprint(value)
	if re.MatchString(str) {
		return nil
	}
	if variable.Validate.Message != "" {
		return fmt.Errorf("%s", variable.Validate.Message)
	}
	return fmt.Errorf("value %q does not match pattern %q", str, pattern)
}

// defaultString renders a manifest default value for display in a text
// prompt, treating a nil default as an empty string (rather than "<nil>").
func defaultString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func parseSelectionValue(value string) any {
	lower := strings.ToLower(value)
	if b, err := strconv.ParseBool(lower); err == nil {
		return b
	}
	return value
}
