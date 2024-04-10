package telegram

import (
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
	f.Core = f.Core.With(fields)
	return f
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
	// TODO add fields in message
	f.TelegramClient.Send(e.Message)
	return nil
}
