package binance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

type CoinPrice struct {
	Coin      string
	AltCoin   string
	Price     decimal.Decimal
	Timestamp time.Time
}

type CoinPriceGroupByAlt struct {
	Coin   string
	Prices []CoinPrice
}

// TODO On devrait peut etre la stocker en DB ?
var SymbolsBlackList = make(map[string]bool)

func (c *Client) GetCoinsPrice(ctx context.Context, coins, altCoins []string) (map[string]CoinPrice, error) {

	symbols := getSymbols(coins, altCoins)

	if len(symbols) == 0 {
		return nil, nil
	}

	prices, err := getPrices(ctx, symbols, c)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	var res = make(map[string]CoinPrice)
	for _, price := range prices {
		coin, altCoin, err := util.Unsymbol(price.Symbol, coins, altCoins)
		if err != nil {
			return nil, fmt.Errorf("couldn't unsymbol %s: %w", price.Symbol, err)
		}
		p, err := decimal.NewFromString(price.Price)
		if err != nil {
			return nil, fmt.Errorf("failed parsing price for %s(%s): %w", price.Symbol, price.Price, err)
		}
		res[price.Symbol] = CoinPrice{
			Coin:      coin,
			AltCoin:   altCoin,
			Price:     p,
			Timestamp: now,
		}
	}

	return res, nil
}

// Sorted by price and altcoin name
func (c *Client) GetCoinsPriceGroupByAltCoins(ctx context.Context, coins, altCoins []string) (map[string]CoinPriceGroupByAlt, error) {

	symbols := getSymbols(coins, altCoins)

	if len(symbols) == 0 {
		return nil, nil
	}

	prices, err := getPrices(ctx, symbols, c)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	var res = make(map[string]CoinPriceGroupByAlt)

	balanceGroupByCoin := util.GroupByProperty(prices, func(s *binance.SymbolPrice) string {
		coin, _, _ := util.Unsymbol(s.Symbol, coins, altCoins)
		return coin
	})

	for coin, symbolPrices := range balanceGroupByCoin {
		var resTmp []CoinPrice
		sort.Slice(symbolPrices, func(i, j int) bool {
			return symbolPrices[i].Symbol < symbolPrices[j].Symbol
		})

		for _, price := range symbolPrices {
			_, altCoin, err := util.Unsymbol(price.Symbol, coins, altCoins)
			if err != nil {
				return nil, fmt.Errorf("couldn't unsymbol %s: %w", coin, err)
			}
			p, err := decimal.NewFromString(price.Price)
			if err != nil {
				return nil, fmt.Errorf("failed parsing price for %s(%s): %w", price.Symbol, price.Price, err)
			}
			resTmp = append(resTmp, CoinPrice{
				Coin:      coin,
				AltCoin:   altCoin,
				Price:     p,
				Timestamp: now,
			})
		}
		res[coin] = CoinPriceGroupByAlt{
			Coin:   coin,
			Prices: resTmp,
		}
	}

	return res, nil
}

func getSymbols(coins []string, altCoins []string) map[string]bool {
	var symbols = make(map[string]bool)
	for _, coin := range coins {
		for _, altCoin := range altCoins {
			if coin == altCoin {
				continue
			}
			if symbol := util.Symbol(coin, altCoin); !SymbolsBlackList[symbol] {
				symbols[symbol] = true
			}
		}
	}
	return symbols
}

func getPrices(ctx context.Context, symbols map[string]bool, c *Client) ([]*binance.SymbolPrice, error) {
	uniqueSymbols := util.Keys(symbols)
	prices, err := c.client.NewListPricesService().Symbols(uniqueSymbols).Do(ctx)
	if err != nil {
		if ErrorIs(err, BinanceErrorInvalidSymbol) {
			prices, err = c.dichotomicPriceFetching(ctx, uniqueSymbols)
		}
	}
	return prices, err
}

func (c *Client) GetSymbolPrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
	prices, err := c.client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		return decimal.Zero, err
	}

	if len(prices) == 0 {
		return decimal.Zero, errors.New("no price returned")
	}

	p, err := decimal.NewFromString(prices[0].Price)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed parsing price for %s(%s): %w", symbol, prices[0].Price, err)
	}

	return p, nil
}

func (c *Client) dichotomicPriceFetching(ctx context.Context, symbols []string) ([]*binance.SymbolPrice, error) {

	// TODO Flemme de faire l'algo dichotomic là maintenant lol, pour l'instant je fais du 1 par 1 mais ca va spammer 😬

	var prices []*binance.SymbolPrice

	for _, symbol := range symbols {
		price, err := c.client.NewListPricesService().Symbol(symbol).Do(ctx)
		if err != nil {
			if ErrorIs(err, BinanceErrorInvalidSymbol) {
				c.Logger.Info(fmt.Sprintf("Found unexisting symbol '%s' on Binance, won't fetch it anymore", symbol))
				SymbolsBlackList[symbol] = true
				continue
			}
			return nil, fmt.Errorf("failed to fetch price for symbol %s: %w", symbol, err)
		}
		prices = append(prices, price...)
	}

	return prices, nil
}

func (c *Client) GetSymbolInfos(ctx context.Context, symbol string) (binance.Symbol, error) {
	allInfos := c.coinInfosRefresher.Data(ctx)

	info, ok := allInfos[symbol]
	if !ok {
		return binance.Symbol{}, fmt.Errorf("cannot find symbol infos")
	}

	return info, nil
}

func (c *Client) RefreshSymbolInfos(ctx context.Context) (map[string]binance.Symbol, error) {
	infos, err := c.client.NewExchangeInfoService().Symbols(c.ConfigFile.GenerateAllSymbolsWithBridge()...).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get coin infos: %w", err)
	}

	var res map[string]binance.Symbol = make(map[string]binance.Symbol)
	for _, info := range infos.Symbols {
		res[info.Symbol] = info
	}
	return res, nil
}
