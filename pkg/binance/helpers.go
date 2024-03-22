package binance

import (
	"strings"

	"github.com/shopspring/decimal"
)

// func StepSizePosition(stepSize string) int32 {
// 	if len(strings.Split(stepSize, "1")[0]) == 0 {
// 		return 1 - int32(len(strings.Split(stepSize, ".")[0]))
// 	}
// 	return int32(len(strings.Split(stepSize, "1")))
// }

func StepSizeFormat(val decimal.Decimal, stepSize string) string {
	if step := strings.Index(stepSize, "1") - 1; step > 0 {
		return val.RoundFloor(int32(step)).String()
	} else {
		return val.Floor().String()
	}
}
