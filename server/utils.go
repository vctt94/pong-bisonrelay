package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/decred/slog"
)

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

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}
