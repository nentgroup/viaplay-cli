package git

import "testing"

func TestValidateGitRemote(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		want   bool
	}{
		// Valid remotes
		{"simple origin", "origin", true},
		{"simple upstream", "upstream", true},
		{"with dash", "upstream-1", true},
		{"with dot", "feature.branch", true},
		{"with underscore", "A_B", true},
		{"mixed allowed specials", "A_B.c-1", true},
		{"all allowed specials", "Z9-._", true},

		// Invalid because of dangerous characters
		{"semicolon command injection", "origin;rm -rf /", false},
		{"space in name", "name with space", false},
		{"double ampersand", "name&&bad", false},
		{"double pipe", "bad||name", false},
		{"greater-than", "bad>name", false},
		{"less-than", "<bad", false},
		{"backtick", "`bad`", false},
		{"dollar sign", "$bad", false},
		{"backslash", "bad\\name", false},
		{"double quotes", "\"bad\"", false},
		{"single quotes", "'bad'", false},

		// Edge cases
		{"empty string", "", true}, // no characters to violate rules
		{"single allowed char", "a", true},
		{"single disallowed char", "!", false},
		{"non-ascii char", "origin✓", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateGitRemote(tt.remote)
			if got != tt.want {
				t.Errorf("validateGitRemote(%q) = %v, want %v", tt.remote, got, tt.want)
			}
		})
	}
}
