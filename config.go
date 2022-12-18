package main

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	AddCommand    string        `env:"ADD_COMMAND" envDefault:"ekle"`
	DoneMessage   string        `env:"DONE_MESSAGE" envDefault:"anlaşıldı"`
	RemoveCommand string        `env:"REMOVE_COMMAND" envDefault:"sil"`
	DatabaseFile  string        `env:"DATABASE_FILE" envDefault:"data.db"`
	FetchDelay    time.Duration `env:"FETCH_DELAY" envDefault:"10s"`
	BotToken      string        `env:"BOT_TOKEN"`
}

func NewConfig() *Config {
	cfg := &Config{}
	if parseErr := env.Parse(cfg); parseErr != nil {
		panic(parseErr)
	}

	return cfg
}
