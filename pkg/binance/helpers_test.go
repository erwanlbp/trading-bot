package binance_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/erwanlbp/trading-bot/pkg/binance"
)

func TestStepSizePosition(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name      string
		inputStep string
		inputVal  string
		expected  string
	}{
		{
			name:      "NEAR",
			inputStep: "0.01000000",
			inputVal:  "18584.63172",
			expected:  "18584.63",
		},
	} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.expected, binance.StepSizeFormat(decimal.RequireFromString(c.inputVal), c.inputStep))
		})
	}
}
