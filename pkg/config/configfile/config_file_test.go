package configfile_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
)

func TestGetNeededGain(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name     string
		input    configfile.Jump
		lastJump time.Time
		expected decimal.Decimal
	}{
		{
			name: "no jump",
			input: configfile.Jump{
				WhenGain:        decimal.NewFromInt(1),
				DecreaseBy:      decimal.NewFromFloat(0.5),
				After:           1 * time.Minute,
				Min:             decimal.NewFromFloat(0.1),
				DefaultLastJump: time.Now().Add(-65 * time.Second),
			},
			expected: decimal.NewFromFloat(0.005),
		},
		{
			name: "no decrease",
			input: configfile.Jump{
				WhenGain:   decimal.NewFromInt(1),
				DecreaseBy: decimal.NewFromFloat(0.5),
				After:      1 * time.Minute,
				Min:        decimal.NewFromFloat(0.1),
			},
			lastJump: time.Now().Add(-1 * time.Second),
			expected: decimal.NewFromFloat(0.01),
		},
		{
			name: "one decrease",
			input: configfile.Jump{
				WhenGain:   decimal.NewFromInt(1),
				DecreaseBy: decimal.NewFromFloat(0.3),
				After:      2 * time.Minute,
				Min:        decimal.NewFromFloat(0.1),
			},
			lastJump: time.Now().Add(-3 * time.Minute),
			expected: decimal.NewFromFloat(0.007),
		},
		{
			name: "multiple decrease",
			input: configfile.Jump{
				WhenGain:   decimal.NewFromInt(1),
				DecreaseBy: decimal.NewFromFloat(0.2),
				After:      2 * time.Minute,
				Min:        decimal.NewFromFloat(0.1),
			},
			lastJump: time.Now().Add(-5 * time.Minute),
			expected: decimal.NewFromFloat(0.006),
		},
		{
			name: "too many decreases",
			input: configfile.Jump{
				WhenGain:   decimal.NewFromInt(1),
				DecreaseBy: decimal.NewFromFloat(0.2),
				After:      2 * time.Minute,
				Min:        decimal.NewFromFloat(0.1),
			},
			lastJump: time.Now().Add(-1 * time.Hour),
			expected: decimal.NewFromFloat(0.001),
		},
		{
			name: "gain equals min",
			input: configfile.Jump{
				WhenGain:   decimal.NewFromInt(1),
				DecreaseBy: decimal.NewFromFloat(0.2),
				After:      2 * time.Minute,
				Min:        decimal.NewFromFloat(1),
			},
			lastJump: time.Now().Add(-1 * time.Hour),
			expected: decimal.NewFromFloat(0.01),
		},
	} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := c.input.GetNeededGain(c.lastJump)

			assert.Equal(t, c.expected.String(), actual.String())
		})
	}
}
