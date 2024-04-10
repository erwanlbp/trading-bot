package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/eventbus/eventdefinition"
)

type ZapCoreWrapper func(zapcore.Core) zapcore.Core

type Logger struct {
	*zap.Logger
	EventBus *eventbus.Bus
}

func NewSimpleZapLogger(e *eventbus.Bus) *Logger {
	l := &Logger{
		Bus: e,
	}
	l.Init(nil)
	return l
}

func NewZapLogger(e *eventbus.Bus, telegramZapCoreWrapper ZapCoreWrapper) *Logger {
	l := &Logger{
		EventBus: e,
	}
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

func (l *Logger) With(fields ...zapcore.Field) *Logger {
	if len(fields) == 0 {
		return l
	}

	log := *l
	log.Core().With(fields)
	return &log
}

func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) DebugWithNotif(msg string, fields ...zapcore.Field) {
	l.Bus.Notify(eventbus.GenerateEvent(eventbus.SendNotification, eventdefinition.EventNotification{Level: eventdefinition.MINOR, Message: msg}))
	l.Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) WarnWithNotif(msg string, fields ...zapcore.Field) {
	l.Bus.Notify(eventbus.GenerateEvent(eventbus.SendNotification, eventdefinition.EventNotification{Level: eventdefinition.MEDIUM, Message: msg}))
	l.Warn(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) InfoWithNotif(msg string, fields ...zapcore.Field) {
	l.Bus.Notify(eventbus.GenerateEvent(eventbus.SendNotification, eventdefinition.EventNotification{Level: eventdefinition.MEDIUM, Message: msg}))
	l.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zapcore.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) ErrorWithNotif(msg string, fields ...zapcore.Field) {
	l.Bus.Notify(eventbus.GenerateEvent(eventbus.SendNotification, eventdefinition.EventNotification{Level: eventdefinition.MAJOR, Message: msg}))
	l.Error(msg, fields...)
}
