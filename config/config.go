package config

import (
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	SlackSigningSecret string `split_words:"true"`
	SlackToken         string `split_words:"true"`
	GoogleAPIToken     string `split_words:"true" default:"key.json"`
}

func NewEnvConfig() (*EnvConfig, error) {
	var c EnvConfig
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
