package config

import (
	"os"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/db"
	"github.com/erwanlbp/trading-bot/pkg/db/sqlite"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/process"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/service"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
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

	ProcessPriceGetter  *process.PriceGetter
	ProcessJumpFinder   *process.JumpFinder
	ProcessFeeGetter    *process.FeeGetter
	ProcessNotification *process.Notification

	NotificationLevel []string
}

func Init() *Config {

	var conf Config

	conf.EventBus = eventbus.NewEventBus()

	conf.Logger = log.NewZapLogger(conf.EventBus)

	cf, err := configfile.ParseConfigFile()
	if err != nil {
		conf.Logger.Fatal("Failed to parse config file", zap.Error(err))
	}
	conf.ConfigFile = &cf

	conf.BinanceClient = binance.NewClient(conf.Logger, conf.ConfigFile, cf.Binance.APIKey, cf.Binance.APIKeySecret)

	telebot, err := telegram.NewClient(conf.Logger, cf.Telegram.Token, cf.Telegram.ChannelId)
	if err != nil {
		conf.Logger.Warn("Failed to init telegram bot (trading-bot sill running)", zap.Error(err))
	}
	conf.TelegramClient = telebot

	dbFolderName := "data"
	dbFileName := "trading_bot"
	if conf.ConfigFile.TestMode {
		dbFileName = "test_trading_bot"
	}
	if rootPath, ok := os.LookupEnv("ROOT_PATH"); ok {
		dbFolderName = rootPath + dbFolderName
	}
	sqliteDb, err := sqlite.NewDB(conf.Logger, dbFolderName, dbFileName)
	if err != nil {
		conf.Logger.Fatal("Failed to initialize DB", zap.Error(err))
	}
	conf.DB = db.NewDB(sqliteDb)

	conf.Repository = repository.NewRepository(conf.DB, conf.ConfigFile, conf.Logger)

	conf.Service = service.NewService(conf.Logger, conf.Repository, conf.BinanceClient, conf.ConfigFile)

	conf.ProcessPriceGetter = process.NewPriceGetter(conf.Logger, conf.BinanceClient, conf.Repository, conf.EventBus, AltCoins)
	conf.ProcessJumpFinder = process.NewJumpFinder(conf.Logger, conf.Repository, conf.EventBus, conf.ConfigFile, conf.BinanceClient)
	conf.ProcessFeeGetter = process.NewFeeGetter(conf.Logger, conf.BinanceClient)
	conf.ProcessNotification = process.NewNotification(conf.Logger, conf.EventBus, conf.TelegramClient)

	return &conf
}
