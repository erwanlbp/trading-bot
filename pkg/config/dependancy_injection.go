package config

import (
	"github.com/erwanlbp/trading-bot/pkg/db"
	"github.com/erwanlbp/trading-bot/pkg/db/sqlite"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"go.uber.org/zap"
)

type Config struct {
	Logger     *log.Logger
	DB         *db.DB
	Repository *repository.Repository
}

func Load() *Config {

	var conf Config

	conf.Logger = log.NewZapLogger()

	dbFileName := "data/trading_bot" // TODO Get it more dynamically ?
	sqliteDb, err := sqlite.NewDB(conf.Logger, dbFileName)
	if err != nil {
		conf.Logger.Fatal("Failed to initialize DB", zap.Error(err))
	}
	conf.DB = db.NewDB(sqliteDb)

	conf.Repository = repository.NewRepository(conf.DB)

	return &conf
}
