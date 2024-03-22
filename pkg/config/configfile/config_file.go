package configfile

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
	yaml "gopkg.in/yaml.v3"
)

type ConfigFile struct {
	TestMode bool `yaml:"test_mode"`

	Binance struct {
		APIKey       string `yaml:"api_key"`
		APIKeySecret string `yaml:"api_key_secret"`
		Tld          string `yaml:"tld"`
	} `yaml:"binance"`
	Bridge string   `yaml:"bridge"`
	Coins  []string `yaml:"coins"`

	// TODO Do we need it ? we could find the ratio getting better and buy it
	StartCoin *string `yaml:"start_coin"`

	TradeTimeout time.Duration `yaml:"trade_timeout"`

	Jump Jump `yaml:"jump"`
}

type Jump struct {
	WhenGain   decimal.Decimal `yaml:"when_gain"`
	DecreaseBy decimal.Decimal `yaml:"decrease_by"`
	After      time.Duration   `yaml:"after"`
	Min        decimal.Decimal `yaml:"min"`

	// Will contains bot start time
	DefaultLastJump time.Time `yaml:"-"`
}

// Return needed ratio (between 0 and 1)
func (j Jump) GetNeededGain(lastJump time.Time) decimal.Decimal {
	gain := j.WhenGain

	if lastJump.IsZero() {
		lastJump = j.DefaultLastJump
	}

	t := time.Now()
	for {
		t = t.Add(-j.After)
		if lastJump.Before(t) {
			gain = gain.Sub(j.DecreaseBy)
			if gain.LessThanOrEqual(j.Min) {
				gain = j.Min
				break
			}
		} else {
			break
		}
	}

	return gain.Div(decimal.NewFromInt(100))
}

func (cf ConfigFile) GenerateAllSymbolsWithBridge() []string {
	var res []string
	for _, coin := range cf.Coins {
		res = append(res, util.Symbol(coin, cf.Bridge))
	}
	return res
}

func (cf *ConfigFile) ApplyDefaults() {
	if cf.TradeTimeout == 0 {
		cf.TradeTimeout = 10 * time.Minute
	}

	// TODO other defaults
}

func ParseConfigFile() (ConfigFile, error) {
	var res ConfigFile

	filepath := "config/config.yaml" // TODO Get it more dynamically ?
	if rootPath, ok := os.LookupEnv("ROOT_PATH"); ok {
		filepath = rootPath + filepath
	}
	file, err := os.Open(filepath)
	if err != nil {
		return res, fmt.Errorf("failed opening file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return res, fmt.Errorf("failed reading file: %w", err)
	}

	var data ConfigFile
	if err := yaml.Unmarshal(content, &data); err != nil {
		return res, fmt.Errorf("failed unmarshaling file: %w", err)
	}

	// To debug if the config is correctly parsed
	// yamled, _ := yaml.Marshal(data)
	// fmt.Print(string(yamled))

	res.ApplyDefaults()

	data.Jump.DefaultLastJump = time.Now()

	return data, nil
}
