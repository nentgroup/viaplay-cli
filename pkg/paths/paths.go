package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Expand expands the tilde in path to the user's home directory
func Expand(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // Return original if home can't be determined
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
