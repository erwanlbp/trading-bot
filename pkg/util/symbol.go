package util

import (
	"fmt"
)

func Symbol(coin, altCoin string) string {
	return fmt.Sprintf("%s%s", coin, altCoin)
}

func Unsymbol(symbol string, referentialCoins, altCoins []string) (string, string, error) {
	for _, coin := range referentialCoins {
		for _, altCoin := range altCoins {
			if Symbol(coin, altCoin) == symbol {
				return coin, altCoin, nil
			}
			if Symbol(altCoin, coin) == symbol {
				return altCoin, coin, nil
			}
		}
	}
	return "", "", ErrNotFound
}
