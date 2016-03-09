package main

import (
	"github.com/BurntSushi/toml"
)

type CollectorConfig struct {
	Address       string
	Port          string
	LogLevel      string
	File          fileConfig
	Elasticsearch esConfig
	Kafka         kafkaConfig
}

type fileConfig struct {
	Enabled bool
	Path    string
	Format  string
}

type esConfig struct {
	Enabled bool
	Host    string
	Index   string
}

type kafkaConfig struct {
	Enabled bool
	Brokers []string
	Topic   string
}

func ReadConfig(file string) (*CollectorConfig, error) {
	var conf = CollectorConfig{
		Address:  "",
		Port:     "8080",
		LogLevel: "info",
		File: fileConfig{
			Enabled: false,
			Path:    ".",
			Format:  "xml",
		},
		Elasticsearch: esConfig{
			Enabled: false,
			Host:    "localhost",
			Index:   "gracc",
		},
		Kafka: kafkaConfig{
			Enabled: false,
			Brokers: []string{"localhost:9092"},
			Topic:   "gracc",
		},
	}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
