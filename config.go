package main

import (
	"time"

	"github.com/BurntSushi/toml"
)

type CollectorConfig struct {
	Address  string
	Port     string
	Timeout  time.Duration
	LogLevel string
	AMQP     AMQPConfig
}

func ReadConfig(file string) (*CollectorConfig, error) {
	var conf = CollectorConfig{
		Address:  "",
		Port:     "8080",
		Timeout:  60,
		LogLevel: "info",
		AMQP: AMQPConfig{
			Host:         "localhost",
			Port:         "5672",
			Format:       "raw",
			Exchange:     "",
			ExchangeType: "fanout",
			Durable:      false,
			AutoDelete:   true,
			Internal:     false,
			RoutingKey:   "",
			Retry:        10,
		},
	}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}
	conf.Timeout = conf.Timeout * time.Second
	conf.AMQP.Retry = conf.AMQP.Retry * time.Second
	return &conf, nil
}
