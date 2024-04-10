package telegram

import (
	"fmt"
	"strings"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Wrapper around the ZapLogger Core.
// We just want to catch messages to send them to Telegram,
// so we forward all the logic to the wrapped Core
type TelegramZapCore struct {
	zapcore.Core

	TelegramClient *Client
	ConfigFile     *configfile.ConfigFile
}

var _ zapcore.Core = &TelegramZapCore{}

func ZapCoreWrapper(tc *Client, cf *configfile.ConfigFile) func(core zapcore.Core) zapcore.Core {
	return func(core zapcore.Core) zapcore.Core {
		return &TelegramZapCore{
			Core:           core,
			TelegramClient: tc,
			ConfigFile:     cf,
		}
	}
}

// With adds structured context to the Core.
func (f *TelegramZapCore) With(fields []zap.Field) zapcore.Core {
	newCore := *f
	newCore.Core = newCore.Core.With(fields)
	return &newCore
}

// Check determines whether the supplied Entry should be logged (using the
// embedded LevelEnabler and possibly some extra logic). If the entry
// should be logged, the Core adds itself to the CheckedEntry and returns
// the result.
//
// Callers must use Check before calling Write.
func (f *TelegramZapCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	res := f.Core.Check(e, ce)

	level, err := zapcore.ParseLevel(f.ConfigFile.NotificationLevel)
	if err != nil {
		return res
	}

	if level.Enabled(e.Level) {
		res.AddCore(e, f)
	}

	return res
}

// Write serializes the Entry and any Fields supplied at the log site and
// writes them to their destination.
//
// If called, Write should always log the Entry and Fields; it should not
// replicate the logic of Check.
func (f *TelegramZapCore) Write(e zapcore.Entry, fields []zap.Field) error {

	message := getIcon(e.Level) + e.Message

	if len(fields) > 0 {
		var fieldsStr []string
		for _, field := range fields {
			switch field.Type {
			case zapcore.ErrorType:
				fieldsStr = append(fieldsStr, field.Key+": "+field.Interface.(error).Error())
			default:
				if field.String != "" {
					fieldsStr = append(fieldsStr, field.Key+": "+field.String)
				} else if field.Interface != nil {
					fieldsStr = append(fieldsStr, field.Key+": "+fmt.Sprintf("%+v", field.Interface))
				}
			}
		}

		message = message + "\n```\n" + strings.Join(fieldsStr, "\n") + "\n```"
	}

	f.TelegramClient.Send(message)
	return nil
}

func getIcon(lvl zapcore.Level) string {
	switch lvl {
	case zapcore.DebugLevel:
		return "🐞 "
	case zapcore.WarnLevel:
		return "️⚠️ "
	case zapcore.ErrorLevel:
		return "🚨 "
	case zapcore.FatalLevel, zapcore.PanicLevel:
		return "💥 "
	default:
		return ""
	}
}
