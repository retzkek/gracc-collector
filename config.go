package main

import (
	"github.com/BurntSushi/toml"
)

type kafkaConfig struct {
	Brokers []string
	Topic   string
}

type config struct {
	Address  string
	Port     string
	Kafka    kafkaConfig
	LogLevel string
}

func ReadConfig(file string) (*config, error) {
	var conf = config{
		Address: "",
		Port:    "8080",
		Kafka: kafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "gratia",
		},
		LogLevel: "info",
	}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
