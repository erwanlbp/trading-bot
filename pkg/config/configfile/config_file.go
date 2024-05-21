package configfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/erwanlbp/trading-bot/pkg/config/globalconf"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
	yaml "gopkg.in/yaml.v3"
)

type ConfigFile struct {
	TestMode bool `yaml:"test_mode"`

	Binance struct {
		APIKey       string `yaml:"api_key,omitempty"`
		APIKeySecret string `yaml:"api_key_secret,omitempty"`
	} `yaml:"binance"`
	Bridge string   `yaml:"bridge"`
	Coins  []string `yaml:"coins"`

	// TODO Do we need it ? we could find the ratio getting better and buy it
	StartCoin *string `yaml:"start_coin"`

	TradeTimeout time.Duration `yaml:"trade_timeout"`

	Jump Jump `yaml:"jump"`

	Order struct {
		Refresh time.Duration `yaml:"refresh"`
	} `yaml:"order"`

	Telegram struct {
		Token     string `yaml:"token,omitempty"`
		ChannelID string `yaml:"channel_id,omitempty"`
		Handlers  struct {
			NbDiffDisplayed int `yaml:"nb_diff_displayed"`
		} `yaml:"handlers"`
	} `yaml:"telegram"`

	NotificationLevel string `yaml:"notification_level"`
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

	t := util.Now().UTC()
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
	var res map[string]bool = make(map[string]bool)
	for _, coin := range cf.Coins {
		res[util.Symbol(coin, cf.Bridge)] = true
	}
	return util.Keys(res)
}

func (cf *ConfigFile) ApplyDefaults() {
	if cf.TradeTimeout == 0 {
		cf.TradeTimeout = 10 * time.Minute
	}
	if cf.Order.Refresh == 0 {
		cf.Order.Refresh = 15 * time.Second
	}
	if len(cf.NotificationLevel) == 0 {
		cf.NotificationLevel = zapcore.InfoLevel.String()
	}

	// TODO other defaults
}

func getConfigFilePath() string {
	filename := "config"

	if globalconf.IsBacktesting() {
		filename = "backtesting_config"
	}

	filepath := fmt.Sprintf("config/%s.yaml", filename)
	if rootPath, ok := os.LookupEnv("ROOT_PATH"); ok {
		filepath = rootPath + filepath
	}
	return filepath
}

func ParseConfigFile() (ConfigFile, error) {
	var res ConfigFile

	filepath := getConfigFilePath()
	file, err := os.Open(filepath)
	if err != nil {
		return res, fmt.Errorf("failed opening file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return res, fmt.Errorf("failed reading file: %w", err)
	}

	if err := yaml.Unmarshal(content, &res); err != nil {
		return res, fmt.Errorf("failed unmarshaling file: %w", err)
	}

	res.ApplyDefaults()

	// To debug if the config is correctly parsed
	// yamled, _ := yaml.Marshal(res)
	// fmt.Print(string(yamled))

	res.Jump.DefaultLastJump = util.Now().UTC()

	return res, nil
}

func CopyFileToBackup() error {
	src := getConfigFilePath()
	dst := src + ".bkp"

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed reading file '%s': %w", src, err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed writing file '%s': %w", dst, err)
	}

	return nil
}

func (c *ConfigFile) SaveToFile() error {

	content, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal conf object to yaml: %w", err)
	}

	filepath := getConfigFilePath()
	if err := os.WriteFile(filepath, content, 0644); err != nil {
		return fmt.Errorf("failed writing file '%s': %w", filepath, err)
	}

	return nil
}

func (nc *ConfigFile) ValidateChanges(pc ConfigFile) error {
	if nc.TestMode != pc.TestMode {
		return errors.New("cannot change test_mode")
	}
	if nc.Binance != pc.Binance {
		return errors.New("cannot change object binance")
	}
	// Maybe we could allow it but I'm not sure of the impacts 😬
	if nc.Bridge != pc.Bridge {
		return errors.New("cannot change bridge")
	}

	// Keep DefaultLastJump date as the original bot start date
	nc.Jump.DefaultLastJump = pc.Jump.DefaultLastJump
	return nil
}

func (c *ConfigFile) RemoveSecrets() {
	c.Binance.APIKey = "<hidden>"
	c.Binance.APIKeySecret = "<hidden>"
	c.Telegram.Token = "<hidden>"
	c.Telegram.ChannelID = "<hidden>"
}
