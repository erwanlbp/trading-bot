package handlers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) Configuration(ctx context.Context) {
	p.Notification(ctx)
	p.TelegramClient.CreateHandler(&btnReloadConfig, p.ReloadConfigFile)
	p.TelegramClient.CreateHandler(&btnShowLiveConfig, p.ShowLiveConfig)
	p.TelegramClient.CreateHandler(&btnShowConfigFile, p.ShowConfigFile)
	p.TelegramClient.CreateHandler(&btnListCoins, p.ListCoins)
	p.TelegramClient.CreateHandler(&btnEditCoins, p.EditCoins)
	p.TelegramClient.CreateHandler("/edit_coins", p.ValidateCoinEdit)

	p.TelegramClient.CreateHandler(&btnConfiguration, func(c telebot.Context) error {
		return c.Send("What do you want to do ?", configurationMenu)
	})
}

func (p *Handlers) ShowConfigFile(c telebot.Context) error {
	confFile, err := configfile.ParseConfigFile()
	if err != nil {
		return c.Send("Failed to parse config.yaml: " + err.Error())
	}
	return c.Send(PrepareConfContentForMessage(confFile), mainMenu, telebot.ModeMarkdown)
}

func (p *Handlers) ShowLiveConfig(c telebot.Context) error {
	return c.Send(PrepareConfContentForMessage(*p.Conf), mainMenu, telebot.ModeMarkdown)
}

func (p *Handlers) ReloadConfigFile(c telebot.Context) error {
	if err := p.GlobalConf.ReloadConfigFile(context.Background()); err != nil {
		return c.Send(fmt.Sprintf("Failed to reload config file: %s", err.Error()))
	}

	response := "Reloaded config file\n" + PrepareConfContentForMessage(*p.Conf)

	return c.Send(response, mainMenu, telebot.ModeMarkdown)
}

func (p *Handlers) ListCoins(c telebot.Context) error {

	coins, err := p.Repository.GetAllCoins()
	if err != nil {
		return c.Send("Failed to get coins: " + err.Error())
	}

	sort.Slice(coins, func(i, j int) bool {
		return coins[i].Coin < coins[j].Coin
	})

	coinsStr := util.ToASCIITable(coins,
		[]string{"Coin", "Enabled", "Since"},
		func(c model.Coin) []string {
			var e int
			if c.Enabled {
				e = 1
			}
			return []string{c.Coin, strconv.Itoa(e), c.EnabledOn.Format("2006-01-02")}
		})

	var messageParts []string = []string{}
	messageParts = append(messageParts, "```", coinsStr, "```")

	return c.Send(strings.Join(messageParts, "\n"), telebot.ModeMarkdownV2, telebot.RemoveKeyboard, configurationMenu)
}

func (p *Handlers) EditCoins(c telebot.Context) error {
	var messageParts []string = []string{
		"Copy and paste the code",
		"Edit the coins and send it to validate",
		"Or ignore this message to do nothing",
	}

	coinsSlice := append([]string{}, p.Conf.Coins...)
	sort.Strings(coinsSlice)
	messageParts = append(messageParts, "```", "/edit_coins", strings.Join(coinsSlice, " "), "```")

	return c.Send(strings.Join(messageParts, "\n"), telebot.ModeMarkdownV2, telebot.RemoveKeyboard, mainMenu)
}

func (p *Handlers) ValidateCoinEdit(c telebot.Context) error {

	coins := c.Args()

	if err := configfile.CopyFileToBackup(); err != nil {
		return c.Send("Failed to backup the config.yaml: " + err.Error())
	}

	newConf := util.Copy(*p.Conf)
	newConf.Coins = util.Distinct(coins)
	sort.Strings(newConf.Coins)

	if err := newConf.SaveToFile(); err != nil {
		return c.Send("Failed to save conf to file: " + err.Error())
	}

	return c.Send("Saved coin list.\n⚠️*You'll need to reload the config file for it to be effective*⚠️\nNew conf is:\n"+PrepareConfContentForMessage(newConf), telebot.ModeMarkdown, mainMenu)
}

func PrepareConfContentForMessage(conf configfile.ConfigFile) string {
	conf.RemoveSecrets()

	confContent := util.ToYAML(conf)
	return "```\n" + confContent + "\n```"
}
