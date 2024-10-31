package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/decred/slog"
	"google.golang.org/grpc/metadata"
)

func getClientIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata found in context")
	}

	clientIDs, ok := md["client-id"]
	if !ok || len(clientIDs) == 0 {
		return "", fmt.Errorf("client-id not found in metadata")
	}

	return clientIDs[0], nil
}

func generateRandomID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetDebugLevel(debugStr string) slog.Level {
	// Convert debugStr to slog.Level
	var debugLevel slog.Level
	switch debugStr {
	case "info":
		debugLevel = slog.LevelInfo
	case "warn":
		debugLevel = slog.LevelWarn
	case "error":
		debugLevel = slog.LevelError
	case "debug":
		debugLevel = slog.LevelDebug
	default:
		log.Fatalf("Unknown debug level: %s", debugStr)
	}

	return debugLevel
}
