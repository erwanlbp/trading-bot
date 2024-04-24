package process

import (
	"context"
	"fmt"

	"github.com/prprprus/scheduler"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/config/globalconf"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type Cleaner struct {
	Logger     *log.Logger
	Repository *repository.Repository
	GlobalConf globalconf.GlobalConfModifier
}

func NewCleaner(l *log.Logger, r *repository.Repository, gc globalconf.GlobalConfModifier) *Cleaner {
	return &Cleaner{
		Logger:     l,
		Repository: r,
		GlobalConf: gc,
	}
}

func (p *Cleaner) Start(ctx context.Context) {
	go func() {

		Scheduler, _ := scheduler.NewScheduler(1000)

		var ids []string

		ids = append(ids,
			Scheduler.Every().Hour(6).Minute(15).Second(15).Do(p.CleanPairHistory),
			Scheduler.Every().Hour(6).Minute(0).Second(15).Do(p.CleanCoinPrice),
			Scheduler.Every().Hour(10).Minute(0).Second(15).Do(p.VacuumDB),
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

func (p *Cleaner) CleanPairHistory() {
	if inserted, deleted, err := p.Repository.CleanOldPairHistory(); err != nil {
		p.Logger.Warn("Failed to clean pair history, if you this multiple days in a row, there's something wrong", zap.Error(err))
	} else {
		p.Logger.Debug(fmt.Sprintf("Inserted %d aggregated pair history, deleted %d lines after aggregation", inserted, deleted), zap.String("process", "cleaner"))
	}
}

func (p *Cleaner) CleanCoinPrice() {
	if inserted, deleted, err := p.Repository.CleanCoinPriceHistory(); err != nil {
		p.Logger.Warn("Failed to clean coin price history, if you this multiple days in a row, there's something wrong", zap.Error(err))
	} else {
		p.Logger.Debug(fmt.Sprintf("Inserted %d aggregated coin price, deleted %d lines after aggregation", inserted, deleted), zap.String("process", "cleaner"))
	}
}

func (p *Cleaner) VacuumDB() {
	if err := p.Repository.Vacuum(); err != nil {
		p.Logger.Warn("Failed to vacuum, there's something wrong", zap.Error(err))
	} else {
		var sizeStr string
		size, err := p.GlobalConf.GetDBSize()
		if err != nil {
			sizeStr = fmt.Sprintf("'err:%s'", err.Error())
		} else {
			sizeStr = util.ByteCountSI(size)
		}
		p.Logger.Info(fmt.Sprintf("Vacuumed, DB file size now is %s", sizeStr), zap.String("process", "cleaner"))
	}
}
