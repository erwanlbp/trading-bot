package process

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/constant"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/prprprus/scheduler"
)

type BalanceSaver struct {
	mtx           sync.Mutex
	Logger        *log.Logger
	Repository    *repository.Repository
	EventBus      *eventbus.Bus
	BinanceClient *binance.Client
}

func NewBalanceSaver(l *log.Logger, r *repository.Repository, e *eventbus.Bus, b *binance.Client) *BalanceSaver {
	return &BalanceSaver{
		Logger:        l,
		Repository:    r,
		EventBus:      e,
		BinanceClient: b,
	}
}

func (p *BalanceSaver) Start(ctx context.Context) {
	sub := p.EventBus.Subscribe(eventbus.SaveBalance)
	go sub.Handler(ctx, p.SaveBalanceBus)

	go func() {

		Scheduler, _ := scheduler.NewScheduler(1000)

		var ids []string

		ids = append(ids,
			Scheduler.Every().Hour(12).Minute(0).Second(0).Do(p.SaveBalanceBatch, ctx), // 8AM Europe/Paris
		)

		// If ctx is canceled, we'll stop the job
		<-ctx.Done()

		for _, id := range ids {
			if err := Scheduler.CancelJob(id); err != nil {
				p.Logger.Error("failed canceling job", zap.Error(err))
			}
		}
	}()
}

func (p *BalanceSaver) SaveBalanceBatch(ctx context.Context) {
	p.SaveBalance(ctx)
}

func (p *BalanceSaver) SaveBalanceBus(ctx context.Context, _ eventbus.Event) {
	p.SaveBalance(ctx)
}

func (p *BalanceSaver) SaveBalance(ctx context.Context) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// history, err := p.Repository.GetLastBalanceHistory()
	// if err != nil {
	// 	p.Logger.Error("failed getting last balance history", zap.Error(err))
	// 	return
	// }

	// Don't need to save if we jump last 24 hours
	// TODO Commenting because with this, if we jump multiple times per day, we only save one point, at least for now I want details lol
	// if history.Timestamp.IsZero() || time.Since(history.Timestamp) > util.Day {
	// return
	// }

	value, err := p.BinanceClient.GetBalanceValue(ctx, []string{constant.USDT, constant.BTC})
	if err != nil {
		p.Logger.Error("failed getting balance value", zap.Error(err))
		return
	}

	balanceToSave := model.BalanceHistory{
		BtcBalance:  value[constant.BTC],
		UsdtBalance: value[constant.USDT],
		Timestamp:   time.Now().UTC(),
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, balanceToSave); err != nil {
		p.Logger.Error("failed saving balance", zap.Error(err))
		return
	}

	p.Logger.Info("Saved balance", zap.String("BTC", balanceToSave.BtcBalance.String()), zap.String("USDT", balanceToSave.UsdtBalance.String()))
}
