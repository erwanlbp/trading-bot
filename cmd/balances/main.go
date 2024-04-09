package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conf := config.Init(ctx)

	LogBalances(ctx, conf.BinanceClient, conf.ConfigFile)
	fmt.Println()
	LogCurrentCoin(conf.Repository)
	fmt.Println()
	LogUSDTValue(ctx, conf.BinanceClient, conf.ConfigFile)
	LogDiff(conf.Repository)
}

func LogCurrentCoin(repo *repository.Repository) {
	cc, hasEverJumped, err := repo.GetCurrentCoin()
	if err != nil {
		fmt.Println("Failed to get current coin:", err.Error())
		return
	}
	if cc.Coin != "" {
		fmt.Println("Current coin is", cc.Coin)
		return
	}

	if !hasEverJumped {
		fmt.Println("No current because never jumped")
		return
	}

	fmt.Println("No current but has jumped before ? that's weird")
}

func LogBalances(ctx context.Context, binance *binance.Client, config *configfile.ConfigFile) {
	b, err := binance.GetBalance(ctx, append(config.Coins, config.Bridge)...)
	if err != nil {
		fmt.Println("Failed to get balances:", err.Error())
	} else {
		fmt.Printf("Balances are %s\n", util.ToJSON(b))
	}
}

func LogUSDTValue(ctx context.Context, binance *binance.Client, config *configfile.ConfigFile) {
	b, err := binance.GetBalance(ctx, append(config.Coins, config.Bridge)...)
	if err != nil {
		fmt.Println("Failed to get balances:", err.Error())
		return
	}

	coins := util.Keys(b)
	prices, err := binance.GetCoinsPrice(ctx, coins, []string{config.Bridge})
	if err != nil {
		fmt.Println("Failed to get prices:", err.Error())
		return
	}

	var totalValue decimal.Decimal
	for _, coin := range coins {
		var value decimal.Decimal
		if coin == config.Bridge {
			value = b[coin]
		} else {
			value = b[coin].Mul(prices[util.Symbol(coin, config.Bridge)].Price)
		}
		fmt.Printf("%s: %s %s\n", coin, value, config.Bridge)
		totalValue = totalValue.Add(value)
	}
	fmt.Printf("Total value: %s %s\n\n", totalValue, config.Bridge)
}

func LogDiff(r *repository.Repository) {
	diff, err := r.GetDiff()
	if err != nil {
		fmt.Println("Failed to get diff :", err.Error())
		return
	}

	sort.Slice(diff, func(i, j int) bool {
		return diff[i].Diff.GreaterThan(diff[j].Diff)
	})
	fmt.Printf("Diff informations : %s \n", diff[0].Timestamp)
	for _, m := range diff {
		fmt.Printf("%s -> %s : %.4s \n", m.FromCoin, m.ToCoin, m.Diff)
	}
}
