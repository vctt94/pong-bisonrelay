package main

import (
	"os"
	"path/filepath"
	"strings"
)

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			// Handle error, maybe return the original path or exit
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
