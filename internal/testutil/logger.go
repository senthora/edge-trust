package testutil

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

var debugLogs = flag.Bool("debug-logs", false, "enable test logs")

func NewLogger() *zap.Logger {
	debugEnv := os.Getenv("TEST_DEBUG_LOGS")
	if *debugLogs || debugEnv != "" {
		logger, _ := zap.NewDevelopment()
		return logger
	}
	return zap.NewNop()
}
