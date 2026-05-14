package config

import (
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(debug bool) *zap.Logger {
	config := zap.NewProductionConfig()

	level := zapcore.InfoLevel
	if debug {
		level = zap.DebugLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)
	config.Encoding = "console"
	config.DisableCaller = true
	config.EncoderConfig.ConsoleSeparator = " "

	config.EncoderConfig.EncodeTime = func(
		t time.Time,
		enc zapcore.PrimitiveArrayEncoder,
	) {
		enc.AppendString("[" + t.Format(time.RFC3339) + "]")
	}
	config.EncoderConfig.EncodeLevel = func(
		l zapcore.Level,
		enc zapcore.PrimitiveArrayEncoder,
	) {
		level := strings.ToUpper(l.String())

		if len(level) < 5 {
			level = level + strings.Repeat(" ", 5-len(level))
		}

		enc.AppendString("[" + level + "]")
	}
	logger, err := config.Build(
		zap.AddStacktrace(zapcore.PanicLevel),
	)
	if err != nil {
		panic(err)
	}
	return logger
}
