package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Stage string

var (
	Dev  Stage = "dev"
	Prod Stage = "prod"
)

// New creates a new zap.Logger instance configured for the specified stage (Dev or Prod)
// and log level.
func New(stage Stage, logLevel string) (*zap.Logger, error) {
	var cfg zap.Config

	switch stage {
	case "prod":
		cfg = zap.NewProductionConfig()
	case "dev":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig = zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalColorLevelEncoder,
		LineEnding:  zapcore.DefaultLineEnding,
	}

	level, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("logger init: failed to parse log level: %w", err)
	}
	cfg.Level = level

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("logger init: failed to build logger: %w", err)
	}
	return logger, nil
}
