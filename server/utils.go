package server

import (
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
