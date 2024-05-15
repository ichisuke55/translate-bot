package config

import (
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	SlackBotToken  string `split_words:"true"`
	SlackAppToken  string `split_words:"true"`
	GoogleAPIToken string `split_words:"true" default:"key.json"`
	ProjectID      string `split_words:"true"`
	LogFile        string `split_words:"true" default:"bot-log.log"`
}

func NewEnvConfig() (*EnvConfig, error) {
	var c EnvConfig
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
