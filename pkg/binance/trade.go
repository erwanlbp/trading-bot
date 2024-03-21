package binance

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (c *Client) Sell(ctx context.Context, coin, stableCoin string) (any, error) {
	return c.Trade(ctx, coin, stableCoin, binance.SideTypeSell)
}

func (c *Client) Buy(ctx context.Context, coin, stableCoin string) (any, error) {
	return c.Trade(ctx, coin, stableCoin, binance.SideTypeBuy)
}

func (c *Client) Trade(ctx context.Context, coin, stableCoin string, side binance.SideType) (int64, error) {
	logger := c.Logger.With(zap.Any("trade", side))

	balances, err := c.GetBalance(ctx, coin, stableCoin)
	if err != nil {
		logger.Error("Failed to get coins balance", zap.Error(err), zap.Strings("coins", []string{coin, stableCoin}))
		return 0, err
	}

	var balance decimal.Decimal
	if side == binance.SideTypeBuy {
		balance = decimal.NewFromFloat(balances[stableCoin])
	} else {
		balance = decimal.NewFromFloat(balances[coin])
	}

	symbol := util.Symbol(coin, stableCoin)

	price, err := c.GetSymbolPrice(ctx, symbol)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get symbol '%s' price", symbol), zap.Error(err))
		return 0, err
	}

	symbolInfo, err := c.GetSymbolInfos(ctx, symbol)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get symbol '%s' infos", symbol), zap.Error(err))
		return 0, err
	}

	// Check that we have enough coin
	min, _ := decimal.NewFromString(symbolInfo.NotionalFilter().MinNotional)
	if value := price.Mul(balance); value.LessThan(min) {
		logger.Error("Don't have enough of current coin to make the trade", zap.String("balance", balance.String()), zap.String("price", price.String()), zap.String("price*balance", value.String()), zap.String("min", min.String()))
		return 0, fmt.Errorf("insufficient balance")
	}

	// Calculate how much quantity we can sell
	step, _ := decimal.NewFromString(symbolInfo.LotSizeFilter().StepSize)

	quantity := balance.Mul(step).Floor().Div(step)
	assetPrecision := symbolInfo.BaseAssetPrecision

	orderQuery := c.client.NewCreateOrderService().
		Quantity(quantity.StringFixed(int32(assetPrecision))).
		Price(price.StringFixed(int32(symbolInfo.QuotePrecision))).
		Side(side).
		Symbol(symbol).
		TimeInForce(binance.TimeInForceTypeGTC).
		Type(binance.OrderTypeLimit)

	if c.ConfigFile.TestMode {
		if err := orderQuery.Test(ctx); err != nil {
			return 0, fmt.Errorf("failed to create test order: %w", err)
		}
		return 0, nil
	} else {
		res, err := orderQuery.Do(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to create test order: %w", err)
		}
		return res.OrderID, nil
	}
}
