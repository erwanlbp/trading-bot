package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapCoreWrapper func(zapcore.Core) zapcore.Core

type Logger struct {
	*zap.Logger
}

func NewSimpleZapLogger() *Logger {
	l := &Logger{}
	l.Init(nil)
	return l
}

func NewZapLogger(telegramZapCoreWrapper ZapCoreWrapper) *Logger {
	l := &Logger{}
	l.Init(telegramZapCoreWrapper)
	return l
}

func (l *Logger) Init(telegramZapCoreWrapper ZapCoreWrapper) {

	conf := zap.NewProductionConfig()

	conf.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	conf.Encoding = "console"

	conf.EncoderConfig = zap.NewProductionEncoderConfig()
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	conf.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	conf.OutputPaths = []string{"stdout"}
	conf.ErrorOutputPaths = []string{"stderr"}
	conf.DisableStacktrace = true

	var opts []zap.Option
	if telegramZapCoreWrapper != nil {
		opts = append(opts, zap.WrapCore(telegramZapCoreWrapper))
	}

	logger, err := conf.Build(opts...)
	if err != nil {
		panic(err)
	}
	l.Logger = logger
}
