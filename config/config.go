package config

import (
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	SlackSigningSecret string `split_words:"true"`
	SlackBotToken      string `split_words:"true"`
	SlackAppToken      string `split_words:"true"`
	GoogleAPIToken     string `split_words:"true" default:"key.json"`
	ListenPort         string `split_words:"true" default:"8080"`
}

func NewEnvConfig() (*EnvConfig, error) {
	var c EnvConfig
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
