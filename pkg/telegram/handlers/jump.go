package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) NextJump(c telebot.Context) error {

	cc, _, err := p.Repository.GetCurrentCoin()
	if err != nil {
		return c.Send("Failed getting current coin: " + err.Error())
	}

	diffs, err := p.Repository.GetDiff(repository.FromCoin(cc.Coin), repository.OrderBy("diff desc"), repository.Limit(p.Conf.Telegram.Handlers.NbDiffDisplayed))
	if err != nil {
		return c.Send("Error while getting next jump info, please retry: " + err.Error())
	}
	if len(diffs) == 0 {
		return c.Send("No diff found")
	}

	var ts string
	msg := util.ToASCIITable(diffs, []string{"Pair", "Ratio diff"}, nil, func(diff model.Diff) []string {
		ts = fmt.Sprintf("Jump at : %s\nNeeds gain of %s\n", diff.Timestamp.Format(time.DateTime), diff.NeededDiff.Mul(decimal.NewFromInt(100)).StringFixed(1))
		return []string{diff.LogSymbol(), diff.Diff.Mul(decimal.NewFromInt(100)).StringFixed(1) + " %"}
	})

	parts := []string{ts, telegram.FormatForMD(msg)}

	return c.Send(strings.Join(parts, "\n"))
}

func (p *Handlers) BestJump(c telebot.Context) error {

	diffs, err := p.Repository.GetDiff(repository.OrderBy("diff desc"), repository.Limit(p.Conf.Telegram.Handlers.NbDiffDisplayed))
	if err != nil {
		return c.Send("Error while getting best jumps: " + err.Error())
	}
	if len(diffs) == 0 {
		return c.Send("No diff found")
	}

	var ts string
	msg := util.ToASCIITable(diffs, []string{"Pair", "Ratio diff"}, nil, func(diff model.Diff) []string {
		ts = fmt.Sprintf("Best jump at : %s\nNeeds gain of %s\n", diff.Timestamp.Format(time.DateTime), diff.NeededDiff.Mul(decimal.NewFromInt(100)).StringFixed(1))
		return []string{diff.LogSymbol(), diff.Diff.Mul(decimal.NewFromInt(100)).StringFixed(1) + " %"}
	})

	parts := []string{ts, telegram.FormatForMD(msg)}

	return c.Send(strings.Join(parts, "\n"))
}

func (p *Handlers) LastTenJumps(c telebot.Context) error {
	jumps, err := p.Repository.GetJumps(repository.OrderBy("timestamp desc"), repository.Limit(10))
	if err != nil {
		return c.Send("Error while getting last ten jump, please retry")
	}
	if len(jumps) < 1 {
		return c.Send("No jump found in DB")
	}

	msg := util.ToASCIITable(jumps, []string{"Date", "Pair"}, nil, func(jump model.Jump) []string {
		return []string{
			jump.Timestamp.Format(time.DateOnly) + "\n" + jump.Timestamp.Format(time.TimeOnly),
			util.LogSymbol(jump.FromCoin, jump.ToCoin),
		}
	})

	return c.Send(telegram.FormatForMD(msg))
}

func (p *Handlers) EditJump(c telebot.Context) error {
	var messageParts []string = []string{
		"Copy and paste the command",
		"Edit the config and send it to validate",
		"Or ignore this message to do nothing",
	}

	jump := p.Conf.Jump

	messageParts = append(messageParts, fmt.Sprintf(
		"`/edit_jump when:%s decrease:%s after:%s min:%s`",
		jump.WhenGain, jump.DecreaseBy, jump.After, jump.Min,
	))

	return c.Send(strings.Join(messageParts, "\n"), telebot.RemoveKeyboard, configurationMenu)
}

func (p *Handlers) ValidateJumpEdit(c telebot.Context) error {

	var jumpConf configfile.Jump = util.Copy(p.Conf.Jump)

	parts := c.Args()

	for _, part := range parts {
		splitted := strings.Split(part, ":")
		if len(splitted) != 2 {
			continue
		}
		switch arg := splitted[1]; splitted[0] {
		case "when":
			val, err := decimal.NewFromString(arg)
			if err != nil {
				return c.Send(fmt.Sprintf("couldn't parse 'when' (%s) argument: %s", arg, err.Error()))
			}
			jumpConf.WhenGain = val
		case "decrease":
			val, err := decimal.NewFromString(arg)
			if err != nil {
				return c.Send(fmt.Sprintf("couldn't parse 'decrease' (%s) argument: %s", arg, err.Error()))
			}
			jumpConf.DecreaseBy = val
		case "after":
			val, err := time.ParseDuration(arg)
			if err != nil {
				return c.Send(fmt.Sprintf("couldn't parse 'after' (%s) argument: %s", arg, err.Error()))
			}
			jumpConf.After = val
		case "min":
			val, err := decimal.NewFromString(arg)
			if err != nil {
				return c.Send(fmt.Sprintf("couldn't parse 'min' (%s) argument: %s", arg, err.Error()))
			}
			jumpConf.Min = val
		default:
			continue
		}
	}

	configFile, err := configfile.ParseConfigFile()
	if err != nil {
		return c.Send(fmt.Sprintf("Failed to parse config.yaml: %s", err.Error()))
	}

	if jumpConf == configFile.Jump {
		return c.Send("Nothing to change with config.yaml file content")
	}

	if err := configfile.CopyFileToBackup(); err != nil {
		return c.Send("Failed to backup the config.yaml: " + err.Error())
	}

	newConf := util.Copy(*p.Conf)
	newConf.Jump = jumpConf

	if err := newConf.SaveToFile(); err != nil {
		return c.Send("Failed to save conf to file: " + err.Error())
	}

	return c.Send("Saved jump config.\n⚠️*You'll need to reload the config file for it to be effective*⚠️\nNew conf is:\n"+PrepareConfContentForMessage(newConf), mainMenu)
}
