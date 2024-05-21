package binance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (c *Client) Sell(ctx context.Context, coin, stableCoin string) (OrderResult, error) {
	return c.Trade(ctx, coin, stableCoin, binance.SideTypeSell)
}

func (c *Client) Buy(ctx context.Context, coin, stableCoin string) (OrderResult, error) {
	return c.Trade(ctx, coin, stableCoin, binance.SideTypeBuy)
}

// return an error if a trade is in progress, otherwise return a release func to call when trade is over.
//
// ⚠️ Don't go concurrently too hard on this func, it's not concurrent safe, but that should be ok for our needs
func (c *Client) TradeLock() (func(), error) {
	if c.tradeInProgress.Load() {
		return nil, fmt.Errorf("Trade is in progress")
	}
	c.tradeInProgress.Store(true)
	return func() {
		c.tradeInProgress.Store(false)
	}, nil
}

func (c *Client) IsTradeInProgress() bool {
	return c.tradeInProgress.Load()
}

// Do not call this one directly, use .Buy() or .Sell()
func (c *Client) Trade(ctx context.Context, coin, stableCoin string, side binance.SideType) (OrderResult, error) {
	logger := c.Logger.With(zap.Any("trade", side))

	balances, err := c.GetBalance(ctx)
	if err != nil {
		logger.Error("Failed to get coins balance", zap.Error(err), zap.Strings("coins", []string{coin, stableCoin}))
		return OrderResult{}, err
	}

	var balance decimal.Decimal
	if side == binance.SideTypeBuy {
		balance = balances[stableCoin]
	} else {
		balance = balances[coin]
	}

	symbol := util.Symbol(coin, stableCoin)

	price, err := c.GetSymbolPrice(ctx, symbol)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get symbol '%s' price", symbol), zap.Error(err))
		return OrderResult{}, err
	}

	symbolInfo, err := c.GetSymbolInfos(ctx, symbol)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get symbol '%s' infos", symbol), zap.Error(err))
		return OrderResult{}, err
	}

	// Calculate how much quantity we can sell
	step := symbolInfo.LotSizeFilter().StepSize

	var quantityBeforeStepSizeFloor decimal.Decimal
	if side == binance.SideTypeBuy {
		quantityBeforeStepSizeFloor = balance.Div(price)
	} else {
		quantityBeforeStepSizeFloor = balance
	}
	quantity := StepSizeFormat(quantityBeforeStepSizeFloor, step)

	// TODO That'd be cool to log the dust, but my formula seems not good ...
	if side == binance.SideTypeBuy {
		logger.Info(fmt.Sprintf("I have %s %s and %s %s. I'll buy %s %s, at price %s", balances[coin], coin, balances[stableCoin], stableCoin, quantity, coin, price), zap.String("step", step))
	} else {
		logger.Info(fmt.Sprintf("I have %s %s and %s %s. I'll sell %s %s, at price %s", balances[coin], coin, balances[stableCoin], stableCoin, quantity, coin, price), zap.String("step", step))
	}

	res, err := c.client.NewCreateOrderService().
		Quantity(quantity).
		Price(price.StringFixed(int32(symbolInfo.QuotePrecision))).
		Side(side).
		Symbol(symbol).
		TimeInForce(binance.TimeInForceTypeGTC).
		Type(binance.OrderTypeLimit).
		Do(ctx)
	if err != nil {
		return OrderResult{}, fmt.Errorf("failed to create order: %w", err)
	}

	order, err := c.WaitForOrderCompletion(ctx, symbol, res.OrderID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to wait for order '%d' completion", res.OrderID), zap.Error(err))
		return OrderResult{}, err
	}

	return order, nil
}

func (c *Client) WaitForOrderCompletion(ctx context.Context, symbol string, orderId int64) (OrderResult, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, c.ConfigFile.TradeTimeout)
	defer cancel()

	ticker := time.NewTicker(c.ConfigFile.Order.Refresh)

	var orderLastStatus *binance.Order
	for {
		select {
		case <-ctx.Done():
			c.Logger.Debug(fmt.Sprintf("Context is done while waiting for order completion, canceling order '%d'", orderId), zap.String("last_status", string(orderLastStatus.Status)))

			// Context that is not canceled yet, that'll have only 1.5s to cancel the order, before the bot is forced turned off
			cancelCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second+500*time.Millisecond)
			defer cancel()

			cancelStatus, err := c.client.NewCancelOrderService().Symbol(symbol).OrderID(orderId).Do(cancelCtx)
			if err != nil {
				c.Logger.Error("Failed to cancel order", zap.Error(err))
				return OrderResult{Order: orderLastStatus, Cancel: cancelStatus}, err
			}

			c.Logger.Info(fmt.Sprintf("Canceled order '%d' because bot is stopping", orderId))

			return OrderResult{Order: orderLastStatus, Cancel: cancelStatus}, errors.New("context canceled")
		case <-timeoutCtx.Done():
			c.Logger.Error("Reached timeout while waiting for order completion, canceling it", zap.String("last_status", string(orderLastStatus.Status)))
			cancelStatus, err := c.client.NewCancelOrderService().Symbol(symbol).OrderID(orderId).Do(ctx)
			if err != nil {
				c.Logger.Error("Failed to cancel order", zap.Error(err))
				return OrderResult{Order: orderLastStatus, Cancel: cancelStatus}, err
			}
			return OrderResult{Order: orderLastStatus, Cancel: cancelStatus}, fmt.Errorf("wait timeout reached")
		case <-ticker.C:
			order, err := c.client.NewGetOrderService().Symbol(symbol).OrderID(orderId).Do(ctx)
			if err != nil {
				c.Logger.Error("Error while waiting for order completion, will continue to wait (and retry) until timeout", zap.Error(err))
				continue
			}
			orderLastStatus = order
			switch order.Status {
			case binance.OrderStatusTypeNew:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is new", order.OrderID))
			case binance.OrderStatusTypePartiallyFilled:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is partially filled (%s/%s)", order.OrderID, order.ExecutedQuantity, order.OrigQuantity))
			case binance.OrderStatusTypeFilled:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is filled", order.OrderID))
				return OrderResult{Order: orderLastStatus}, nil
			case binance.OrderStatusTypeRejected:
				c.Logger.Error(fmt.Sprintf("Order '%d' got rejected", order.OrderID))
				return OrderResult{Order: orderLastStatus}, fmt.Errorf("order got rejected")
			case binance.OrderStatusTypePendingCancel:
				c.Logger.Debug(fmt.Sprintf("Order '%d' is pending cancel", order.OrderID))
			case binance.OrderStatusTypeCanceled:
				c.Logger.Error(fmt.Sprintf("Order '%d' is canceled", order.OrderID))
				return OrderResult{Order: orderLastStatus}, fmt.Errorf("order got canceled")
			case binance.OrderStatusTypeExpired:
				c.Logger.Warn(fmt.Sprintf("Order '%d' is expired", order.OrderID))
				return OrderResult{Order: orderLastStatus}, fmt.Errorf("order is expired")
			default:
				c.Logger.Warn(fmt.Sprintf("Unknown status '%s' while waiting for order completion, will continue to wait", order.Status))
			}
		}
	}
}

type OrderResult struct {
	Order  *binance.Order
	Cancel *binance.CancelOrderResponse
}

func (r OrderResult) IsPartiallyExecuted() bool {
	return (r.Order != nil && r.Order.Status == binance.OrderStatusTypePartiallyFilled) ||
		(r.Cancel != nil && r.Cancel.Status == binance.OrderStatusTypePartiallyFilled)
}

func (r OrderResult) Price() decimal.Decimal {
	if r.Order != nil {
		return decimal.RequireFromString(r.Order.Price)
	}
	if r.Cancel != nil {
		return decimal.RequireFromString(r.Cancel.Price)
	}
	return decimal.Zero
}

func (r OrderResult) Quantity() decimal.Decimal {
	if r.Cancel != nil {
		return decimal.RequireFromString(r.Cancel.ExecutedQuantity)
	}
	if r.Order != nil {
		return decimal.RequireFromString(r.Order.ExecutedQuantity)
	}
	return decimal.Zero
}

func (r OrderResult) Time() time.Time {
	if r.Order != nil {
		return time.UnixMilli(r.Order.Time)
	}
	return time.Time{}
}
