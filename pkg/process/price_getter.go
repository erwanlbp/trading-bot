package process

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type PriceGetter struct {
	Logger        *log.Logger
	BinanceClient *binance.Client
	Repository    *repository.Repository
}

func NewPriceGetter(l *log.Logger, bc *binance.Client, r *repository.Repository) *PriceGetter {
	return &PriceGetter{
		Logger:        l,
		BinanceClient: bc,
		Repository:    r,
	}
}

func (p *PriceGetter) Run(ctx context.Context) error {

	coins, err := p.Repository.GetEnabledCoins()
	if err != nil {
		return fmt.Errorf("failed getting enabled coins: %w", err)
	}

	prices, err := p.BinanceClient.GetCoinsPrice(ctx, coins, "USDT")
	if err != nil {
		return fmt.Errorf("failed getting coins prices: %w", err)
	}

	util.DebugPrintJson(prices)

	return nil
}
