package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func addConf(cfg zap.Config) zap.Config {
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	cfg.Encoding = "console"

	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.DisableStacktrace = true
	return cfg
}
