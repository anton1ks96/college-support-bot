package app

import (
	"github.com/anton1ks96/college-support-bot/internal/bot"
	"github.com/anton1ks96/college-support-bot/internal/config"
	"github.com/anton1ks96/college-support-bot/pkg/logger"
)

func Run() {
	cfg := config.New()

	b, err := bot.New(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	if err := b.Start(); err != nil {
		logger.Fatal(err)
	}
}
