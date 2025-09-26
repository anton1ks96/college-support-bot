package config

import (
	"fmt"

	"github.com/anton1ks96/college-support-bot/pkg/logger"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Error(fmt.Errorf("error loading .env file: %v", err))
	}

}
