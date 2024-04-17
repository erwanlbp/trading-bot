package process

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository"
)

type SymbolBlacklister struct {
	Logger     *log.Logger
	EventBus   *eventbus.Bus
	Repository *repository.Repository

	cache map[string]bool
	mtx   sync.Mutex
}

func NewSymbolBlacklister(l *log.Logger, e *eventbus.Bus, r *repository.Repository) *SymbolBlacklister {
	return &SymbolBlacklister{
		Logger:     l,
		EventBus:   e,
		Repository: r,
	}
}

func (p *SymbolBlacklister) Start(ctx context.Context) {
	sub := p.EventBus.Subscribe(eventbus.EventFoundUnexistingSymbol)

	go sub.Handler(ctx, p.BlacklistSymbol)
}

func (p *SymbolBlacklister) BlacklistSymbol(ctx context.Context, e eventbus.Event) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	symbol := e.Payload.(string)
	if err := p.Repository.BlacklistSymbol(symbol); err != nil {
		p.Logger.Warn(fmt.Sprintf("Failed to blacklist symbol '%s', will do next tick", symbol), zap.Error(err))
		return
	}

	if p.cache == nil {
		if err := p.RefreshCache(); err != nil {
			p.Logger.Warn("failed to refresh blacklisted symbols cache from db, will do next tick", zap.Error(err))
			return
		}
	}

	p.cache[symbol] = true
}

func (p *SymbolBlacklister) IsSymbolBlacklisted(symbol string) bool {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.cache == nil {
		if err := p.RefreshCache(); err != nil {
			p.Logger.Warn("failed to refresh blacklisted symbols cache from db, will do next tick", zap.Error(err))
			return false
		}
	}

	return p.cache[symbol]
}

func (p *SymbolBlacklister) RefreshCache() error {
	symbols, err := p.Repository.GetBlacklistedSymbols()
	if err != nil {
		return err
	}
	p.cache = make(map[string]bool)
	for _, symbol := range symbols {
		p.cache[symbol.Symbol] = true
	}
	return nil
}
