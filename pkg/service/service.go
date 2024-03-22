package service

import (
	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository"
)

type Service struct {
	Logger     *log.Logger
	Repository *repository.Repository
	Binance    *binance.Client
	ConfigFile *configfile.ConfigFile
}

func NewService(l *log.Logger, r *repository.Repository, b *binance.Client, cf *configfile.ConfigFile) *Service {
	return &Service{
		Logger:     l,
		Repository: r,
		Binance:    b,
		ConfigFile: cf,
	}
}
