package config

import (
	"log"
	"shared/config"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	Base          config.Config
	InviteURLBase string `env:"INVITE_URL_BASE,required"`
}

func New() Config {
	var c Config
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return c
}
