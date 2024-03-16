package config

import (
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/db"
	"github.com/erwanlbp/trading-bot/pkg/db/sqlite"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/process"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/service"
)

type Config struct {
	Logger *log.Logger

	ConfigFile ConfigFile

	DB         *db.DB
	Repository *repository.Repository

	Service *service.Service

	EventBus *eventbus.Bus

	BinanceClient *binance.Client

	ProcessPriceGetter *process.PriceGetter
	ProcessPriceLogger *process.PriceLogger
}

func Init() *Config {

	var conf Config

	conf.Logger = log.NewZapLogger()

	cf, err := ParseConfigFile()
	if err != nil {
		conf.Logger.Fatal("Failed to parse config file", zap.Error(err))
	}
	conf.ConfigFile = cf

	conf.BinanceClient = binance.NewClient(cf.Binance.APIKey, cf.Binance.APIKeySecret)

	dbFileName := "data/trading_bot" // TODO Get it more dynamically ?
	sqliteDb, err := sqlite.NewDB(conf.Logger, dbFileName)
	if err != nil {
		conf.Logger.Fatal("Failed to initialize DB", zap.Error(err))
	}
	conf.DB = db.NewDB(sqliteDb)

	conf.Repository = repository.NewRepository(conf.DB)

	conf.EventBus = eventbus.NewEventBus()

	conf.Service = service.NewService(conf.Logger, conf.Repository)

	conf.ProcessPriceGetter = process.NewPriceGetter(conf.Logger, conf.BinanceClient, conf.Repository, conf.EventBus, AltCoins)
	conf.ProcessPriceLogger = process.NewPriceLogger(conf.Logger, conf.Repository, conf.EventBus)

	return &conf
}
