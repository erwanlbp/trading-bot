package process

import (
	"context"
	"time"

	"github.com/prprprus/scheduler"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/log"
)

type FeeGetter struct {
	Logger        *log.Logger
	BinanceClient *binance.Client
}

func NewFeeGetter(l *log.Logger, bc *binance.Client) *FeeGetter {
	return &FeeGetter{
		Logger:        l,
		BinanceClient: bc,
	}
}

func (p *FeeGetter) Start(ctx context.Context) {
	go func() {

		Scheduler, _ := scheduler.NewScheduler(1000)

		id := Scheduler.Every().Second(0).Do(p.FetchBinanceFees, ctx)

		// To avoid waiting too long before first fetch
		if time.Now().Second() < 20 {
			p.FetchBinanceFees(ctx)
		}

		// If ctx is canceled, we'll stop the job
		<-ctx.Done()

		if err := Scheduler.CancelJob(id); err != nil {
			p.Logger.Error("failed canceling job", zap.Error(err))
		}
	}()
}

func (p *FeeGetter) FetchBinanceFees(ctx context.Context) {
	logger := p.Logger.With(zap.String("process", "fetch_fees"))
	logger.Debug("Fetching Binance fees")

	p.BinanceClient.RefreshFees(ctx)
}
