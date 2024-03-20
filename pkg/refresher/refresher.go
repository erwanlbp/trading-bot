package refresher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/erwanlbp/trading-bot/pkg/log"
)

type HandlerFunc[T any] func(context.Context) (T, error)

type Refresher[T any] struct {
	Logger *log.Logger

	data T

	initOnce     sync.Once
	mtx          sync.RWMutex
	singleflight singleflight.Group

	tickerDuration time.Duration
	ticker         *time.Ticker

	refreshFunc HandlerFunc[T]
	onErrorFunc func(error, time.Time)

	lastSuccessful time.Time

	close chan struct{}
}

func NewRefresher[T any](logger *log.Logger, tickerDuration time.Duration, handler HandlerFunc[T], onErrorFunc func(error, time.Time)) *Refresher[T] {
	b := Refresher[T]{
		Logger: logger,

		tickerDuration: tickerDuration,
		ticker:         time.NewTicker(tickerDuration),

		refreshFunc: handler,
		onErrorFunc: onErrorFunc,

		close: make(chan struct{}),
	}
	return &b
}

func (r *Refresher[T]) Data(ctx context.Context) T {
	r.initOnce.Do(func() { r.start(ctx) })

	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.data
}

func (r *Refresher[T]) Stop() {
	if r.ticker != nil {
		r.ticker.Stop()
	}
	if r.close != nil {
		close(r.close)
		r.close = nil
	}
}

func (r *Refresher[T]) start(ctx context.Context) {
	// Trigger at startup to be sure it's loaded
	r.triggerRefresh(ctx)
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.triggerRefresh(ctx)
			case <-ctx.Done():
				r.Stop()
				break
			case <-r.close:
				break
			}
		}
	}()
}

func (r *Refresher[T]) triggerRefresh(ctx context.Context) {
	_, _, _ = r.singleflight.Do("refresh", func() (_ interface{}, _ error) {
		r.mtx.Lock()
		defer r.mtx.Unlock()

		d, err := r.refreshFunc(ctx)
		if err != nil {
			r.onErrorFunc(err, r.lastSuccessful)
			return
		}
		r.data = d
		r.lastSuccessful = time.Now()
		return
	})
}

func OnErrorLog(logger *log.Logger) func(error, time.Time) {
	return func(err error, lastSuccess time.Time) {

		lastSuccessDuration := "never"
		if !lastSuccess.IsZero() {
			lastSuccessDuration = fmt.Sprintf("%s ago", lastSuccess)
		}

		logger.Error("Failed refreshing, last success was "+lastSuccessDuration, zap.Error(err))
	}
}
