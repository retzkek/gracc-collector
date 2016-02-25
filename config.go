package main

import (
	"github.com/BurntSushi/toml"
)

type config struct {
	Address  string
	Port     string
	LogLevel string
	Path     string
	Format   string
}

func ReadConfig(file string) (*config, error) {
	var conf = config{
		Address:  "",
		Port:     "8080",
		LogLevel: "info",
		Path:     ".",
		Format:   "xml",
	}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
