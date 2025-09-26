package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/anton1ks96/college-support-bot/pkg/logger"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		Telegram
	}

	Telegram struct {
		Token string
		Group int64
	}
)

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Error(fmt.Errorf("error loading .env file: %v", err))
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logger.Warn("bot token not set")
	}

	group := os.Getenv("GROUP")
	if group == "" {
		logger.Warn("group not set")
	}
}

func New() *Config {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	groupID, _ := strconv.ParseInt(os.Getenv("GROUP"), 10, 64)

	return &Config{
		Telegram: Telegram{
			Token: token,
			Group: groupID,
		},
	}
}
