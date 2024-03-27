package log

import (
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/eventbus/eventdefinition"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	*eventbus.Bus
}

func NewZapLogger(e *eventbus.Bus) *Logger {
	l := &Logger{
		Bus: e,
	}
	l.Init()
	return l
}

func (l *Logger) Init() {
	conf := addConf(zap.NewProductionConfig())
	l.Logger, _ = conf.Build()
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
