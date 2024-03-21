package binance

import "strings"

func StepSizePosition(stepSize string) int32 {
	if len(strings.Split(stepSize, "1")[0]) == 0 {
		return int32(len(strings.Split(stepSize, ".")[0]))
	}
	return int32(len(strings.Split(stepSize, "1")))
}
