package service

import (
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
)

type Service struct {
	Logger     *log.Logger
	Repository *repository.Repository
}

func NewService(l *log.Logger, r *repository.Repository) *Service {
	return &Service{
		Logger:     l,
		Repository: r,
	}
}
