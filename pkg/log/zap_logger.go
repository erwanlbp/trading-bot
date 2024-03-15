package log

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

func NewZapLogger() *Logger {
	l := &Logger{}
	l.Init()
	return l
}

func (l *Logger) Init() {
	conf := addConf(zap.NewProductionConfig())
	l.Logger, _ = conf.Build()
}
