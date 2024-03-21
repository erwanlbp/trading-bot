package binance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erwanlbp/trading-bot/pkg/binance"
)

func TestStepSizePosition(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name     string
		input    string
		expected int32
	}{
		{
			name:     "AVAX",
			input:    "0.01000000",
			expected: 2,
		},
	} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.expected, binance.StepSizePosition(c.input))
		})
	}
}
