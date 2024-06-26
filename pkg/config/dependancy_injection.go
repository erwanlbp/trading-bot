package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/config/globalconf"
	"github.com/erwanlbp/trading-bot/pkg/constant"
	"github.com/erwanlbp/trading-bot/pkg/db"
	"github.com/erwanlbp/trading-bot/pkg/db/sqlite"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/process"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/service"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/telegram/handlers"
)

type Config struct {
	Logger *log.Logger

	ConfigFile *configfile.ConfigFile

	DB         *db.DB
	Repository *repository.Repository

	Service *service.Service

	EventBus *eventbus.Bus

	BinanceClient  *binance.Client
	TelegramClient *telegram.Client

	ProcessPriceGetter       *process.PriceGetter
	ProcessJumpFinder        *process.JumpFinder
	ProcessFeeGetter         *process.FeeGetter
	ProcessCleaner           *process.Cleaner
	TelegramHandlers         *handlers.Handlers
	ProcessTelegramNotifier  *process.TelegramNotifier
	ProcessSymbolBlacklister *process.SymbolBlacklister
	BalanceSaver             *process.BalanceSaver
}

var _ globalconf.GlobalConfModifier = &Config{}

func Init(ctx context.Context) *Config {

	var conf Config

	conf.EventBus = eventbus.NewEventBus()

	simpleLogger := log.NewSimpleZapLogger()

	cf, err := configfile.ParseConfigFile()
	if err != nil {
		simpleLogger.Fatal("Failed to parse config file", zap.Error(err))
	}
	conf.ConfigFile = &cf

	telebot, err := telegram.NewClient(ctx, simpleLogger, conf.ConfigFile)
	if err != nil {
		simpleLogger.Warn("Failed to init telegram bot (trading-bot still running)", zap.Error(err))
	}
	conf.TelegramClient = telebot

	conf.Logger = log.NewZapLogger(telegram.ZapCoreWrapper(conf.TelegramClient, conf.ConfigFile))

	dbFilePath := getDBFilePath(conf.ConfigFile.TestMode)
	sqliteDb, err := sqlite.NewDB(conf.Logger, dbFilePath)
	if err != nil {
		conf.Logger.Fatal("Failed to initialize DB", zap.Error(err))
	}
	conf.DB = db.NewDB(sqliteDb)

	conf.Repository = repository.NewRepository(conf.DB, conf.ConfigFile, conf.Logger)

	conf.ProcessSymbolBlacklister = process.NewSymbolBlacklister(conf.Logger, conf.EventBus, conf.Repository)

	conf.BinanceClient = binance.NewClient(conf.Logger, conf.ConfigFile, conf.EventBus, conf.ProcessSymbolBlacklister)

	conf.Service = service.NewService(conf.Logger, conf.Repository, conf.BinanceClient, conf.ConfigFile)

	conf.ProcessPriceGetter = process.NewPriceGetter(conf.Logger, conf.BinanceClient, conf.Repository, conf.EventBus, constant.AltCoins)
	conf.ProcessJumpFinder = process.NewJumpFinder(conf.Logger, conf.Repository, conf.EventBus, conf.ConfigFile, conf.BinanceClient)
	conf.ProcessFeeGetter = process.NewFeeGetter(conf.Logger, conf.BinanceClient)
	conf.ProcessCleaner = process.NewCleaner(conf.Logger, conf.Repository, &conf)
	conf.ProcessTelegramNotifier = process.NewTelegramNotifier(conf.Logger, conf.EventBus, conf.TelegramClient)
	conf.TelegramHandlers = handlers.NewHandlers(conf.Logger, conf.ConfigFile, conf.TelegramClient, conf.BinanceClient, conf.Repository, &conf)
	conf.BalanceSaver = process.NewBalanceSaver(conf.Logger, conf.Repository, conf.EventBus, conf.BinanceClient)

	return &conf
}

func getDBFilePath(testMode bool) string {
	filepath := "data/trading_bot.db"
	if testMode {
		filepath = "data/test_trading_bot.db"
	}
	if rootPath, ok := os.LookupEnv("ROOT_PATH"); ok {
		filepath = rootPath + filepath
	}
	return filepath
}

// Re-parse config file, reload enabled coins and stuff. Then replace the ConfigFile in the conf by the new one.
//
// If an errors occurs, the ConfigFile is not replaced, but the DB might have enabled coins, so you must re-re-load the original config to revert
func (c *Config) ReloadConfigFile(ctx context.Context) error {
	logger := c.Logger

	newConfig, err := configfile.ParseConfigFile()
	if err != nil {
		return fmt.Errorf("failed parsing config file: %w", err)
	}

	if err := newConfig.ValidateChanges(*c.ConfigFile); err != nil {
		return fmt.Errorf("invalid live change: %w", err)
	}

	logger.Debug("Reloading supported coins")
	if err := LoadCoins(newConfig.Coins, logger, c.Repository); err != nil {
		return fmt.Errorf("failed to reload supported coins: %w", err)
	}

	logger.Debug("Reloading available pairs")
	if err := c.Service.InitializePairs(ctx); err != nil {
		return fmt.Errorf("failed re-initializing coin pairs: %w", err)
	}

	c.Logger.Debug("Reloading config file")
	*c.ConfigFile = newConfig

	return nil
}

func (c *Config) ExportDBFile() (io.Reader, error) {
	if err := c.Repository.Vacuum(); err != nil {
		return nil, fmt.Errorf("failed to vacuum before export: %w", err)
	}

	content, err := os.ReadFile(getDBFilePath(c.ConfigFile.TestMode))
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(content)
	return b, nil
}

func (c *Config) GetDBSize() (int64, error) {
	stat, err := os.Stat(getDBFilePath(c.ConfigFile.TestMode))
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}
