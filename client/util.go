package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/metadata"
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

func attachClientIDToContext(ctx context.Context, clientID string) context.Context {
	md := metadata.New(map[string]string{
		"client-id": clientID,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
