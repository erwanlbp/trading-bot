package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) NextJump(c telebot.Context) error {
	diffs, err := p.Repository.GetDiff(repository.OrderBy("diff desc"))
	if err != nil {
		return c.Send("Error while get next jump info, please retry")
	}
	if len(diffs) == 0 {
		return c.Send("No diff found")
	}

	chunks := util.Chunk(diffs, p.Conf.Telegram.Handlers.NbDiffDisplayed)
	diffsToPrint := chunks[0]

	var ts string
	msg := util.ToASCIITable(diffsToPrint, []string{"Pair", "Ratio diff"}, nil, func(diff model.Diff) []string {
		ts = fmt.Sprintf("Diff at : %s\nNeeds gain of %s\n", diff.Timestamp.Format(time.DateTime), diff.NeededDiff.Mul(decimal.NewFromInt(100)).StringFixed(1))
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
