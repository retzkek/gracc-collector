package main

import (
	"github.com/BurntSushi/toml"
)

type esConfig struct {
	Host  string
	Index string
}

type config struct {
	Address       string
	Port          string
	Elasticsearch esConfig
	LogLevel      string
}

func ReadConfig(file string) (*config, error) {
	var conf config
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
