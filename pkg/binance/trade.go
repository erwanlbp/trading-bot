package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (c *Client) Sell(ctx context.Context, coin, stableCoin string) (*binance.Order, error) {
	release, err := c.TradeLock()
	if err != nil {
		return nil, err
	}
	defer release()

	return c.Trade(ctx, coin, stableCoin, binance.SideTypeSell)
}

func (c *Client) Buy(ctx context.Context, coin, stableCoin string) (*binance.Order, error) {
	release, err := c.TradeLock()
	if err != nil {
		return nil, err
	}
	defer release()

	return c.Trade(ctx, coin, stableCoin, binance.SideTypeBuy)
}

// return an error if a trade is in progress, otherwise return a release func to call when trade is over.
//
// ⚠️ Don't go concurrently too hard on this func, it's not concurrent safe, but that should be ok for our needs
func (c *Client) TradeLock() (func(), error) {
	if c.tradeInProgress {
		return nil, fmt.Errorf("Trade is in progress")
	}
	c.tradeInProgress = true
	return func() {
		c.tradeInProgress = false
	}, nil
}

func (c *Client) IsTradeInProgress() bool {
	return c.tradeInProgress
}

// Do not call this one directly, use .Buy() or .Sell()
func (c *Client) Trade(ctx context.Context, coin, stableCoin string, side binance.SideType) (*binance.Order, error) {
	logger := c.Logger.With(zap.Any("trade", side))

	balances, err := c.GetBalance(ctx, coin, stableCoin)
	if err != nil {
		logger.Error("Failed to get coins balance", zap.Error(err), zap.Strings("coins", []string{coin, stableCoin}))
		return nil, err
	}

	var (
		balance          decimal.Decimal
		fromCoin, toCoin string
	)
	if side == binance.SideTypeBuy {
		balance = decimal.NewFromFloat(balances[stableCoin])
		fromCoin = stableCoin
		toCoin = coin
	} else {
		balance = decimal.NewFromFloat(balances[coin])
		fromCoin = coin
		toCoin = stableCoin
	}

	symbol := util.Symbol(coin, stableCoin)

	price, err := c.GetSymbolPrice(ctx, symbol)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get symbol '%s' price", symbol), zap.Error(err))
		return nil, err
	}

	symbolInfo, err := c.GetSymbolInfos(ctx, symbol)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get symbol '%s' infos", symbol), zap.Error(err))
		return nil, err
	}

	// Check that we have enough coin
	min := decimal.RequireFromString(symbolInfo.NotionalFilter().MinNotional)
	if value := price.Mul(balance); value.LessThan(min) {
		logger.Error("Don't have enough of current coin to make the trade", zap.String("balance", balance.String()), zap.String("price", price.String()), zap.String("price*balance", value.String()), zap.String("min", min.String()))
		return nil, fmt.Errorf("insufficient balance")
	}

	// Calculate how much quantity we can sell
	step := symbolInfo.LotSizeFilter().StepSize

	assetPrecision := symbolInfo.BaseAssetPrecision
	quantityBeforeStepSizeFloor := balance.Div(price)
	quantity := quantityBeforeStepSizeFloor.RoundFloor(StepSizePosition(step))

	// TODO That'd be cool to log the dust, but my formula seems not good ...
	// logger.Info(fmt.Sprintf("I have %s %s. I'll buy %s %s, at price %s, leaving %s of %s dust", balance, fromCoin, quantity, toCoin, price, quantityBeforeStepSizeFloor.Sub(quantity), fromCoin), zap.String("step", step))
	logger.Info(fmt.Sprintf("I have %s %s. I'll buy %s %s, at price %s", balance, fromCoin, quantity, toCoin, price), zap.String("step", step))

	orderQuery := c.client.NewCreateOrderService().
		Quantity(quantity.StringFixed(int32(assetPrecision))).
		Price(price.StringFixed(int32(symbolInfo.QuotePrecision))).
		Side(side).
		Symbol(symbol).
		TimeInForce(binance.TimeInForceTypeGTC).
		Type(binance.OrderTypeLimit)

	var orderId int64
	if c.ConfigFile.TestMode {
		if err := orderQuery.Test(ctx); err != nil {
			return nil, fmt.Errorf("failed to create test order: %w", err)
		}
		return nil, nil
	} else {
		res, err := orderQuery.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create test order: %w", err)
		}
		orderId = res.OrderID
	}

	order, err := c.WaitForOrderCompletion(ctx, symbol, orderId)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to wait for order '%d' completion", orderId), zap.Error(err))
		return nil, err
	}

	return order, nil
}

func (c *Client) WaitForOrderCompletion(ctx context.Context, symbol string, orderId int64) (*binance.Order, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, c.ConfigFile.TradeTimeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)

	var orderLastStatus binance.OrderStatusType
	for {
		select {
		case <-ctx.Done():
			c.Logger.Debug(fmt.Sprintf("Stopped waiting for order '%d' completion", orderId), zap.String("last_status", string(orderLastStatus)))
			return nil, fmt.Errorf("context canceled")
		case <-timeoutCtx.Done():
			c.Logger.Error("Reached timeout while waiting for order completion, canceling it", zap.String("last_status", string(orderLastStatus)))
			_, err := c.client.NewCancelOrderService().Symbol(symbol).OrderID(orderId).Do(ctx)
			if err != nil {
				c.Logger.Error("Failed to cancel order", zap.Error(err))
				return nil, err
			}
			return nil, fmt.Errorf("wait timeout reached")
		case <-ticker.C:
			order, err := c.client.NewGetOrderService().Symbol(symbol).OrderID(orderId).Do(ctx)
			if err != nil {
				c.Logger.Error("Error while waiting for order completion, will continue to wait (and retry) until timeout", zap.Error(err))
				continue
			}
			orderLastStatus = order.Status
			switch order.Status {
			case binance.OrderStatusTypeNew:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is new", order.OrderID))
			case binance.OrderStatusTypePartiallyFilled:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is partially filled", order.OrderID))
			case binance.OrderStatusTypeFilled:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is filled", order.OrderID))
				return order, nil
			case binance.OrderStatusTypeRejected:
				c.Logger.Error(fmt.Sprintf("Order '%d' got rejected", order.OrderID))
				return nil, fmt.Errorf("order got rejected")
			case binance.OrderStatusTypePendingCancel:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is pending cancel", order.OrderID))
			case binance.OrderStatusTypeCanceled:
				c.Logger.Error(fmt.Sprintf("Order '%d' is canceled", order.OrderID))
				return nil, fmt.Errorf("order got canceled")
			case binance.OrderStatusTypeExpired:
				c.Logger.Warn(fmt.Sprintf("Order '%d' is expired", order.OrderID))
				return nil, fmt.Errorf("order is expired")
			default:
				c.Logger.Warn(fmt.Sprintf("Unknown status '%s' while waiting for order completion, will continue to wait", order.Status))
			}
		}
	}
}
